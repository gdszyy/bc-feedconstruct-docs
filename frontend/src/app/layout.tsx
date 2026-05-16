import type { ReactNode } from "react";

import { Providers } from "./providers";

export const metadata = {
  title: "BC FeedConstruct Web",
  description: "Sports betting data viewer (BFF-driven)",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
