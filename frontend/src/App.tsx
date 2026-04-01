import { useState, useEffect, useRef, CSSProperties } from "react";
import { EventsOn } from "../wailsjs/runtime/runtime";
import { GetCandidates, RegisterToAnki, CheckAnkiConnection, HideWindow, QuitApp } from "../wailsjs/go/main/App";
import CardBuilder from "./components/CardBuilder";

interface SynonymEntry {
  word: string;
  difference: string;
}

interface CandidateSet {
  word: string;
  meanings: string[];
  examples: string[];
  core_image: string;
  synonyms: SynonymEntry[];
  frequency: string;
  priority: string;
  usage: string;
  image_keywords: string[];
  image_urls: string[];
}

const styles: Record<string, CSSProperties> = {
  waiting: {
    height: "100%",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    color: "#999",
    gap: 12,
  },
  loading: {
    height: "100%",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    color: "#666",
    gap: 8,
  },
  error: {
    padding: 24,
    textAlign: "center" as const,
    color: "#c0392b",
  },
  retryBtn: {
    marginTop: 12,
    padding: "8px 16px",
    border: "1px solid #c0392b",
    borderRadius: 6,
    background: "none",
    color: "#c0392b",
    cursor: "pointer",
  },
};

type AppState =
  | { type: "waiting" }
  | { type: "loading"; word: string }
  | { type: "ready"; candidates: CandidateSet }
  | { type: "error"; word: string; message: string }
  | { type: "success" };

function App() {
  const [state, setState] = useState<AppState>({ type: "waiting" });
  const [ankiConnected, setAnkiConnected] = useState<boolean | null>(null);
  const [registering, setRegistering] = useState(false);
  const requestId = useRef(0);

  useEffect(() => {
    CheckAnkiConnection()
      .then(() => setAnkiConnected(true))
      .catch(() => setAnkiConnected(false));
  }, []);

  useEffect(() => {
    const cancel = EventsOn("word:detected", (word: string) => {
      loadCandidates(word);
    });
    return cancel;
  }, []);

  useEffect(() => {
    const cancel = EventsOn("images:loaded", (data: { word: string; image_urls: string[] }) => {
      setState((prev) => {
        if (prev.type === "ready" && prev.candidates.word === data.word) {
          return { ...prev, candidates: { ...prev.candidates, image_urls: data.image_urls } };
        }
        return prev;
      });
    });
    return cancel;
  }, []);

  const loadCandidates = async (word: string) => {
    const currentRequest = ++requestId.current;
    setState({ type: "loading", word });
    try {
      const candidates = await GetCandidates(word);
      if (currentRequest !== requestId.current) return;
      setState({ type: "ready", candidates: candidates as unknown as CandidateSet });
    } catch (err: any) {
      if (currentRequest !== requestId.current) return;
      setState({ type: "error", word, message: err.toString() });
    }
  };

  const handleRegister = async (word: string, meaning: string, example: string, imageUrl: string) => {
    setRegistering(true);
    try {
      await RegisterToAnki({ word, meaning, example, image_url: imageUrl });
      setState({ type: "waiting" });
      HideWindow();
    } catch (err: any) {
      alert("Registration failed: " + err.toString());
    } finally {
      setRegistering(false);
    }
  };

  const handleDismiss = () => {
    setState({ type: "waiting" });
    HideWindow();
  };

  return (
    <div style={{ height: "100%" }}>
      {ankiConnected !== null && (
        <div
          style={{
            position: "fixed" as const, top: 8, right: 12, fontSize: 11,
            padding: "2px 8px", borderRadius: 4,
            backgroundColor: ankiConnected ? "#e8f5e9" : "#fbe9e7",
            color: ankiConnected ? "#2e7d32" : "#c62828",
          }}
        >
          Anki: {ankiConnected ? "Connected" : "Disconnected"}
        </div>
      )}

      {state.type === "waiting" && (
        <div style={styles.waiting}>
          <div style={{ fontSize: 32 }}>Waiting...</div>
          <div>Copy an English word to get started</div>
          <button
            style={{ marginTop: 24, padding: "6px 12px", border: "1px solid #ccc", borderRadius: 4, background: "none", color: "#999", cursor: "pointer", fontSize: 12 }}
            onClick={() => QuitApp()}
          >
            Quit
          </button>
        </div>
      )}

      {state.type === "loading" && (
        <div style={styles.loading}>
          <div style={{ fontSize: 20 }}>Looking up "{state.word}"...</div>
        </div>
      )}

      {state.type === "ready" && (
        <CardBuilder
          candidates={state.candidates}
          onRegister={handleRegister}
          onDismiss={handleDismiss}
          registering={registering}
        />
      )}

      {state.type === "success" && (
        <div style={{
          position: "fixed" as const, top: 0, left: 0, right: 0, bottom: 0,
          display: "flex", alignItems: "center", justifyContent: "center",
          backgroundColor: "rgba(255,255,255,0.9)", fontSize: 18, fontWeight: 600, color: "#27ae60",
        }}>
          Registered!
        </div>
      )}

      {state.type === "error" && (
        <div style={styles.error}>
          <div>Error looking up "{state.word}"</div>
          <div style={{ fontSize: 12, marginTop: 8 }}>{state.message}</div>
          <button style={styles.retryBtn} onClick={() => loadCandidates(state.word)}>
            Retry
          </button>
        </div>
      )}
    </div>
  );
}

export default App;
