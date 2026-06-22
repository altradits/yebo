"use client";
import { InputHTMLAttributes, forwardRef } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
  prefix?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, hint, error, prefix, className = "", id, ...props }, ref) => {
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, "-");
    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label htmlFor={inputId} className="text-sm font-medium text-ink">
            {label}
          </label>
        )}
        <div className="relative flex items-center">
          {prefix && (
            <span className="absolute left-4 text-ink-soft font-medium select-none pointer-events-none">
              {prefix}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            className={[
              "w-full h-14 rounded-btn border bg-surface-card px-4 text-base text-ink",
              "placeholder:text-ink-muted",
              "focus:outline-none focus:ring-2 focus:ring-green-600 focus:border-transparent",
              error ? "border-red-400" : "border-border",
              prefix ? "pl-10" : "",
              className,
            ].join(" ")}
            {...props}
          />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        {hint && !error && <p className="text-sm text-ink-muted">{hint}</p>}
      </div>
    );
  }
);
Input.displayName = "Input";
