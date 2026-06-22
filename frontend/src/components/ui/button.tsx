"use client";
import { ButtonHTMLAttributes, forwardRef } from "react";

type Variant = "primary" | "secondary" | "ghost" | "danger";
type Size    = "lg" | "md" | "sm";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  fullWidth?: boolean;
}

const variants: Record<Variant, string> = {
  primary:   "bg-green-600 text-white active:bg-green-700 disabled:bg-green-600/50",
  secondary: "bg-green-50 text-green-700 border border-green-200 active:bg-green-100",
  ghost:     "bg-transparent text-ink active:bg-border",
  danger:    "bg-red-600 text-white active:bg-red-700",
};

const sizes: Record<Size, string> = {
  lg: "h-14 px-6 text-base font-semibold",
  md: "h-12 px-5 text-sm font-semibold",
  sm: "h-9  px-4 text-sm font-medium",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "lg", loading, fullWidth, className = "", children, disabled, ...props }, ref) => (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={[
        "inline-flex items-center justify-center gap-2 rounded-btn select-none transition-opacity",
        "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-600",
        variants[variant],
        sizes[size],
        fullWidth ? "w-full" : "",
        disabled || loading ? "opacity-60 cursor-not-allowed" : "cursor-pointer",
        className,
      ].join(" ")}
      {...props}
    >
      {loading ? <Spinner /> : children}
    </button>
  )
);
Button.displayName = "Button";

function Spinner() {
  return (
    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  );
}
