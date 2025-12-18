import { useEffect, useRef, useState } from "react";
import { Login } from "../wailsjs/go/api/Api";
import { EventsOn } from "../wailsjs/runtime/runtime";
import QRCode from "qrcode";
import { ChatListScreen } from "./screens/ChatScreen";
import { LoginScreen } from "./screens/LoginScreen";

type Screen = "login" | "chats";

function App() {
  const [screen, setScreen] = useState<Screen>("login");
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [status, setStatus] = useState<string>("waiting");

  useEffect(() => {
    Login();
  }, []);

  useEffect(() => {
    const unsubQR = EventsOn("wa:qr", async (qr: string) => {
      if (!canvasRef.current) return;
      await QRCode.toCanvas(canvasRef.current, qr, { width: 300 });
    });

    const unsubStatus = EventsOn("wa:status", (status: string) => {
      setStatus(status);
      if (status === "logged_in" || status === "success") {
        setScreen("chats");
      }
    });

    return () => {
      unsubQR();
      unsubStatus();
    };
  }, []);

  return (
    <>
      {screen === "login" && (
        <LoginScreen
          canvasRef={canvasRef}
          status={status}
        />
      )}

      {screen === "chats" && <ChatListScreen />}
    </>
  );
}

export default App;
