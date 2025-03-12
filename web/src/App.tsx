import { MouseEvent, useEffect, useRef, useState } from "react";

const socket = new WebSocket("ws://localhost:8080/ws");

interface CursorData {
  [username: string]: {
    x: number,
    y: number
  };
}

export default function Editor() {
  const [username, setUsername] = useState<string>("");
  const [content, setContent] = useState<string>("");
  const [cursors, setCursors] = useState<CursorData>({});
  const editorRef = useRef<HTMLTextAreaElement | null>(null);

  console.log({ content, cursors })

  const connect = () => {
    socket.onopen = () => {
      socket.send(JSON.stringify({ type: "username", username }));
    };
  }

  useEffect(() => {
    socket.onmessage = (event: MessageEvent) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "text") {
        setContent(msg.content);
      } else if (msg.type === "cursor") {
        setCursors((prev) => ({ ...prev, [msg.username]: msg.cursorPos }));
      }
    };
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newText = e.target.value;
    setContent(newText);
    socket.send(JSON.stringify({ type: "text", username, content: newText }));
  };

  const handleMouseMove = (event: MouseEvent<HTMLDivElement | HTMLTextAreaElement>): void => {
    const x = event.clientX;
    const y = event.clientY;

    const cursorPos = { x, y }

    socket.send(JSON.stringify({ type: "cursor", username, cursorPos }));
  };


  return (
    <div>

      <div>
        <input
          type="text"
          placeholder="Enter your name"
          onChange={(e) => setUsername(e.target.value)}
        />
        <button onClick={connect}>
          start
        </button>
      </div>
      
      <h2>Editing as: {username}</h2>
      <textarea
        ref={editorRef}
        value={content}
        onChange={handleChange}
        onMouseMove={handleMouseMove}
        rows={50}
        cols={100}
      />
      <div>
        {Object.entries(cursors).map(([user, pos]) => {
          return <div
            key={user}
            className="custom-cursor"
            style={{
              position: "absolute",
              left: pos.x,
              top: pos.y,
              width: "10px",
              height: "10px",
              backgroundColor: "blue",
              borderRadius: "50%",
              pointerEvents: "none",
              transform: "translate(-50%, -50%)",
            }}
          >{user}</div>
        }
        )}
      </div>
    </div>
  );
}