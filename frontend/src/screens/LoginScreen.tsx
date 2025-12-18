export function LoginScreen({
    canvasRef,
    status,
  }: {
    canvasRef: React.RefObject<HTMLCanvasElement>;
    status?: string;
  }) {
    return (
        <div className="wa-root">
          <div className="wa-card">
            {/* LEFT */}
            <div className="wa-left">
              <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/6/6b/WhatsApp.svg/1022px-WhatsApp.svg.png" className="wa-logo" />
              <h1>Log in to WhatsApp</h1>
    
              <ol>
                <li>Open WhatsApp on your phone</li>
                <li>Tap <b>Menu</b> or <b>Settings</b></li>
                <li>Select <b>Linked Devices</b></li>
                <li>Point your phone at this screen</li>
              </ol>
            </div>
    
            {/* RIGHT */}
            <div className="wa-right">
              <div className="qr-box">
                <canvas ref={canvasRef} />
              </div>
    
              <p className="qr-text">
                Scan this QR code with the WhatsApp app
              </p>
    
              <p className="qr-status">{status}</p>
            </div>
          </div>
        </div>
    );
  }
  