import { Suspense } from "react";
import LoginPage from "./LoginForm";

export default function Page() {
  return (
    <Suspense
      fallback={
        <div style={{ minHeight: "100vh", display: "grid", placeItems: "center" }}>
          Loading…
        </div>
      }
    >
      <LoginPage />
    </Suspense>
  );
}
