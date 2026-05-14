export default function HomePage() {
  return (
    <main style={{ padding: 24, fontFamily: "system-ui" }}>
      <h1>BC FeedConstruct Web — BDD scaffold</h1>
      <p>
        This is the Next.js BFF consumer. The page intentionally renders no
        live data: implementation begins after BDD empty tests are confirmed
        per <code>CLAUDE.md</code>.
      </p>
      <ul>
        <li>BFF HTTP: <code>{process.env.NEXT_PUBLIC_BFF_HTTP ?? "(unset)"}</code></li>
        <li>BFF WS: <code>{process.env.NEXT_PUBLIC_BFF_WS ?? "(unset)"}</code></li>
      </ul>
    </main>
  );
}
