"use client";

import type { ReactNode } from "react";

import { StoresProvider } from "@/react/StoresProvider";

export function Providers({ children }: { children: ReactNode }) {
  return <StoresProvider>{children}</StoresProvider>;
}
