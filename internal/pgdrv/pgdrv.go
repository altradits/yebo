// Package pgdrv implements a zero-dependency PostgreSQL v3 wire protocol driver
// compatible with database/sql/driver. Uses only stdlib: crypto, encoding, net.
//
// Supported: simple queries, parameterised queries (extended protocol),
// SCRAM-SHA-256 and MD5 authentication, TLS via sslmode=require.
package pgdrv

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func init() {
	sql.Register("pgdrv", &Driver{})
}

// Driver implements driver.Driver.
type Driver struct{}

func (d *Driver) Open(dsn string) (driver.Conn, error) { return openConn(dsn) }

// ── DSN ───────────────────────────────────────────────────────────────────────

type config struct{ host, port, dbname, user, password, sslmode string }

func parseDSN(dsn string) config {
	c := config{host: "localhost", port: "5432", sslmode: "disable"}
	for _, part := range strings.Fields(dsn) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		v := strings.Trim(kv[1], "'")
		switch kv[0] {
		case "host":
			c.host = v
		case "port":
			c.port = v
		case "dbname":
			c.dbname = v
		case "user":
			c.user = v
		case "password":
			c.password = v
		case "sslmode":
			c.sslmode = v
		}
	}
	return c
}

// ── conn ──────────────────────────────────────────────────────────────────────

type conn struct {
	nc  net.Conn
	cfg config
}

func openConn(dsn string) (*conn, error) {
	cfg := parseDSN(dsn)
	nc, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.host, cfg.port), 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("pgdrv: dial: %w", err)
	}
	c := &conn{nc: nc, cfg: cfg}
	if cfg.sslmode != "disable" {
		if err := c.upgradeSSL(); err != nil {
			nc.Close()
			return nil, err
		}
	}
	if err := c.startup(); err != nil {
		nc.Close()
		return nil, err
	}
	return c, nil
}

func (c *conn) upgradeSSL() error {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint32(msg[0:4], 8)
	binary.BigEndian.PutUint32(msg[4:8], 80877103)
	if _, err := c.nc.Write(msg); err != nil {
		return err
	}
	resp := make([]byte, 1)
	if _, err := io.ReadFull(c.nc, resp); err != nil {
		return err
	}
	if resp[0] != 'S' {
		return fmt.Errorf("pgdrv: server declined SSL")
	}
	c.nc = tls.Client(c.nc, &tls.Config{ServerName: c.cfg.host})
	return nil
}

func (c *conn) startup() error {
	var params []byte
	for _, kv := range [][2]string{
		{"user", c.cfg.user}, {"database", c.cfg.dbname},
		{"client_encoding", "UTF8"}, {"application_name", "yebobank"},
	} {
		params = append(params, []byte(kv[0]+"\x00"+kv[1]+"\x00")...)
	}
	params = append(params, 0)
	body := append([]byte{0, 3, 0, 0}, params...)
	msg := make([]byte, 4)
	binary.BigEndian.PutUint32(msg, uint32(len(body)+4))
	if _, err := c.nc.Write(append(msg, body...)); err != nil {
		return err
	}
	return c.authLoop()
}

func (c *conn) authLoop() error {
	for {
		typ, data, err := c.readMsg()
		if err != nil {
			return err
		}
		switch typ {
		case 'R':
			if err := c.handleAuth(data); err != nil {
				return err
			}
		case 'Z':
			return nil
		case 'E':
			return parseErr(data)
		}
	}
}

func (c *conn) handleAuth(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("pgdrv: short auth message")
	}
	switch binary.BigEndian.Uint32(data[0:4]) {
	case 0:
		return nil
	case 5:
		if len(data) < 8 {
			return fmt.Errorf("pgdrv: short MD5 salt")
		}
		return c.sendMD5(data[4:8])
	case 10:
		return c.doSCRAM(data[4:])
	default:
		return fmt.Errorf("pgdrv: unsupported auth method %d", binary.BigEndian.Uint32(data[0:4]))
	}
}

func (c *conn) sendMD5(salt []byte) error {
	inner := md5hex([]byte(c.cfg.password + c.cfg.user))
	return c.sendMsg('p', []byte("md5"+md5hex(append([]byte(inner), salt...))+"\x00"))
}

func md5hex(b []byte) string { s := md5.Sum(b); return hex.EncodeToString(s[:]) } // #nosec G401

func (c *conn) doSCRAM(mechs []byte) error {
	found := false
	for _, m := range strings.Split(strings.TrimRight(string(mechs), "\x00"), "\x00") {
		if m == "SCRAM-SHA-256" {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("pgdrv: server does not support SCRAM-SHA-256")
	}
	nonce := make([]byte, 18)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	cn := base64.StdEncoding.EncodeToString(nonce)
	cf := "n,,n=,r=" + cn
	mech := "SCRAM-SHA-256\x00"
	cfb := []byte(cf)
	body := make([]byte, len(mech)+4+len(cfb))
	copy(body, mech)
	binary.BigEndian.PutUint32(body[len(mech):], uint32(len(cfb)))
	copy(body[len(mech)+4:], cfb)
	if err := c.sendMsg('p', body); err != nil {
		return err
	}
	typ, data, err := c.readMsg()
	if err != nil {
		return err
	}
	if typ != 'R' || len(data) < 4 || binary.BigEndian.Uint32(data[0:4]) != 11 {
		return fmt.Errorf("pgdrv: expected SASLContinue")
	}
	sf := parseKV(string(data[4:]))
	sn, s64, si := sf["r"], sf["s"], sf["i"]
	if !strings.HasPrefix(sn, cn) {
		return fmt.Errorf("pgdrv: server nonce mismatch")
	}
	salt, _ := base64.StdEncoding.DecodeString(s64)
	iters, _ := strconv.Atoi(si)
	if iters < 4096 {
		return fmt.Errorf("pgdrv: iteration count too low")
	}
	cfnp := "c=biws,r=" + sn
	am := "n=,r=" + cn + "," + string(data[4:]) + "," + cfnp
	spw := pbkdf2([]byte(c.cfg.password), salt, iters, 32)
	ck := hmacSHA256(spw, []byte("Client Key"))
	sk := sha256.Sum256(ck)
	cs := hmacSHA256(sk[:], []byte(am))
	cp := xor(ck, cs)
	final := []byte(cfnp + ",p=" + base64.StdEncoding.EncodeToString(cp))
	if err := c.sendMsg('p', final); err != nil {
		return err
	}
	_, _, err = c.readMsg()
	return err
}

func pbkdf2(pw, salt []byte, iter, kl int) []byte {
	prf := hmac.New(sha256.New, pw)
	hl := prf.Size()
	nb := (kl + hl - 1) / hl
	dk := make([]byte, 0, nb*hl)
	U := make([]byte, hl)
	for b := 1; b <= nb; b++ {
		prf.Reset()
		prf.Write(salt)
		prf.Write([]byte{0, 0, 0, byte(b)})
		U = prf.Sum(U[:0])
		T := append([]byte{}, U...)
		for i := 1; i < iter; i++ {
			prf.Reset()
			prf.Write(U)
			U = prf.Sum(U[:0])
			for j := range T {
				T[j] ^= U[j]
			}
		}
		dk = append(dk, T...)
	}
	return dk[:kl]
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
func xor(a, b []byte) []byte {
	o := make([]byte, len(a))
	for i := range a {
		o[i] = a[i] ^ b[i]
	}
	return o
}
func parseKV(s string) map[string]string {
	m := make(map[string]string)
	for _, p := range strings.Split(s, ",") {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// ── Wire ──────────────────────────────────────────────────────────────────────

func (c *conn) readMsg() (byte, []byte, error) {
	hdr := make([]byte, 5)
	if _, err := io.ReadFull(c.nc, hdr); err != nil {
		return 0, nil, err
	}
	n := binary.BigEndian.Uint32(hdr[1:5])
	if n < 4 {
		return 0, nil, fmt.Errorf("pgdrv: bad length %d", n)
	}
	body := make([]byte, n-4)
	if _, err := io.ReadFull(c.nc, body); err != nil {
		return 0, nil, err
	}
	return hdr[0], body, nil
}

func (c *conn) sendMsg(typ byte, body []byte) error {
	msg := make([]byte, 5+len(body))
	msg[0] = typ
	binary.BigEndian.PutUint32(msg[1:5], uint32(4+len(body)))
	copy(msg[5:], body)
	_, err := c.nc.Write(msg)
	return err
}

// ── driver.Conn ───────────────────────────────────────────────────────────────

func (c *conn) Prepare(query string) (driver.Stmt, error) { return &stmt{c: c, q: query}, nil }
func (c *conn) Close() error                              { _ = c.sendMsg('X', nil); return c.nc.Close() }
func (c *conn) Begin() (driver.Tx, error) {
	if err := c.simpleExec("BEGIN"); err != nil {
		return nil, err
	}
	return &tx{c}, nil
}

type tx struct{ c *conn }

func (t *tx) Commit() error   { return t.c.simpleExec("COMMIT") }
func (t *tx) Rollback() error { return t.c.simpleExec("ROLLBACK") }

// ── stmt ──────────────────────────────────────────────────────────────────────

type stmt struct {
	c *conn
	q string
}

func (s *stmt) Close() error  { return nil }
func (s *stmt) NumInput() int { return -1 }
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	r, err := s.c.runQuery(s.q, args)
	if err != nil {
		return nil, err
	}
	return driver.RowsAffected(r.affected), nil
}
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) { return s.c.runQuery(s.q, args) }

// ── query engine ──────────────────────────────────────────────────────────────

func (c *conn) simpleExec(sql string) error {
	r, err := c.runQuery(sql, nil)
	if err != nil {
		return err
	}
	r.data = nil
	return nil
}

func (c *conn) runQuery(sql string, args []driver.Value) (*rows, error) {
	if len(args) > 0 {
		return c.extQuery(sql, args)
	}
	return c.simpleQuery(sql)
}

func (c *conn) simpleQuery(sql string) (*rows, error) {
	if err := c.sendMsg('Q', []byte(sql+"\x00")); err != nil {
		return nil, err
	}
	return c.collect()
}

func (c *conn) extQuery(sql string, args []driver.Value) (*rows, error) {
	// Parse
	pb := append([]byte("\x00"), append([]byte(sql+"\x00"), 0, 0)...)
	if err := c.sendMsg('P', pb); err != nil {
		return nil, err
	}
	// Bind
	bb := []byte("\x00\x00\x00\x00")
	bb = append(bb, byte(0), byte(len(args)))
	for _, a := range args {
		if a == nil {
			bb = append(bb, 0xff, 0xff, 0xff, 0xff)
		} else {
			s := []byte(valStr(a))
			bb = appendU32(bb, uint32(len(s)))
			bb = append(bb, s...)
		}
	}
	bb = append(bb, 0, 0)
	if err := c.sendMsg('B', bb); err != nil {
		return nil, err
	}
	// Describe portal — required to receive RowDescription before data rows
	if err := c.sendMsg('D', []byte("P\x00")); err != nil {
		return nil, err
	}
	// Execute + Sync
	if err := c.sendMsg('E', append([]byte("\x00"), 0, 0, 0, 0)); err != nil {
		return nil, err
	}
	if err := c.sendMsg('S', nil); err != nil {
		return nil, err
	}
	return c.collect()
}

func (c *conn) collect() (*rows, error) {
	r := &rows{}
	for {
		typ, data, err := c.readMsg()
		if err != nil {
			return nil, err
		}
		switch typ {
		case 'T':
			r.cols = parseRowDesc(data)
		case 'D':
			r.data = append(r.data, parseDataRow(data))
		case 'C':
			r.affected = parseTag(data)
		case 'Z':
			return r, nil
		case 'E':
			return nil, parseErr(data)
		}
	}
}

// ── rows ──────────────────────────────────────────────────────────────────────

type rows struct {
	cols     []string
	data     [][][]byte
	pos      int
	affected int64
}

func (r *rows) Columns() []string { return r.cols }
func (r *rows) Close() error      { r.data = nil; return nil }
func (r *rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.pos]
	r.pos++
	for i := range dest {
		if i < len(row) && row[i] != nil {
			dest[i] = string(row[i])
		} else {
			dest[i] = nil
		}
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseRowDesc(d []byte) []string {
	if len(d) < 2 {
		return nil
	}
	n := int(binary.BigEndian.Uint16(d[0:2]))
	cols := make([]string, 0, n)
	pos := 2
	for i := 0; i < n && pos < len(d); i++ {
		end := pos
		for end < len(d) && d[end] != 0 {
			end++
		}
		cols = append(cols, string(d[pos:end]))
		pos = end + 19 // null byte + 18 bytes of field metadata
	}
	return cols
}

func parseDataRow(d []byte) [][]byte {
	if len(d) < 2 {
		return nil
	}
	n := int(binary.BigEndian.Uint16(d[0:2]))
	vals := make([][]byte, n)
	pos := 2
	for i := 0; i < n && pos+4 <= len(d); i++ {
		l := int(int32(binary.BigEndian.Uint32(d[pos : pos+4])))
		pos += 4
		if l < 0 {
			continue
		}
		vals[i] = d[pos : pos+l]
		pos += l
	}
	return vals
}

func parseTag(d []byte) int64 {
	s := strings.TrimRight(string(d), "\x00")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}
	n, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return n
}

func parseErr(d []byte) error {
	f := make(map[byte]string)
	for i := 0; i < len(d); {
		if d[i] == 0 {
			break
		}
		code := d[i]
		i++
		end := i
		for end < len(d) && d[end] != 0 {
			end++
		}
		f[code] = string(d[i:end])
		i = end + 1
	}
	return fmt.Errorf("pgdrv: %s: %s", f['C'], f['M'])
}

func valStr(v driver.Value) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case time.Time:
		return t.UTC().Format(time.RFC3339Nano)
	}
	return fmt.Sprint(v)
}

func appendU32(b []byte, v uint32) []byte {
	return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
