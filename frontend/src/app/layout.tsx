import type { ReactNode } from "react";

export const metadata = {
  title: "BC FeedConstruct Web",
  description: "Sports betting data viewer (BFF-driven)",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
