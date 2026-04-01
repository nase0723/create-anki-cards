import { useState, CSSProperties } from "react";
import EditableItem from "./EditableItem";

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

interface CardBuilderProps {
  candidates: CandidateSet;
  onRegister: (word: string, meaning: string, example: string, imageUrls: string[]) => void;
  onDismiss: () => void;
  onRefreshImages: () => void;
  refreshing: boolean;
  registering: boolean;
}

const styles: Record<string, CSSProperties> = {
  container: {
    padding: 24,
    height: "100%",
    display: "flex",
    flexDirection: "column",
  },
  word: {
    fontSize: 28,
    fontWeight: 700,
    textAlign: "center" as const,
    marginBottom: 20,
    color: "#1a1a1a",
  },
  content: {
    flex: 1,
    overflowY: "auto" as const,
  },
  section: {
    marginBottom: 16,
  },
  sectionLabel: {
    fontSize: 12,
    fontWeight: 600,
    color: "#666",
    textTransform: "uppercase" as const,
    letterSpacing: 0.5,
    marginBottom: 6,
    display: "block",
  },
  coreImage: {
    padding: "10px 12px",
    backgroundColor: "#f0f7ff",
    borderRadius: 6,
    borderLeft: "3px solid #4a90d9",
    fontSize: 14,
    lineHeight: 1.6,
    color: "#333",
  },
  listItem: {
    padding: "6px 12px",
    borderRadius: 6,
    border: "1px solid #e0e0e0",
    marginBottom: 4,
    fontSize: 14,
    lineHeight: 1.5,
  },
  synonyms: {
    display: "flex",
    gap: 8,
    flexWrap: "wrap" as const,
  },
  synonymTag: {
    padding: "4px 12px",
    backgroundColor: "#f0f0f0",
    borderRadius: 16,
    fontSize: 13,
    color: "#555",
  },
  meta: {
    display: "flex",
    flexDirection: "column" as const,
    alignItems: "center",
    gap: 6,
    marginBottom: 20,
  },
  badge: {
    padding: "3px 10px",
    borderRadius: 12,
    fontSize: 12,
    fontWeight: 600,
  },
  actions: {
    display: "flex",
    gap: 8,
    marginTop: 16,
    paddingTop: 16,
    borderTop: "1px solid #e0e0e0",
  },
  closeBtn: {
    flex: 1,
    padding: "12px 16px",
    backgroundColor: "#f0f0f0",
    color: "#666",
    border: "none",
    borderRadius: 8,
    fontSize: 15,
    cursor: "pointer",
  },
};

function FrequencyGauge({ level }: { level: string }) {
  const levels: Record<string, number> = {
    "Very Common": 4,
    "Common": 3,
    "Uncommon": 2,
    "Rare": 1,
  };
  const filled = levels[level] ?? 2;
  const colors = ["#e53935", "#ff9800", "#8bc34a", "#4caf50"];

  return (
    <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
      <span style={{ fontSize: 12, color: "#666", minWidth: 65 }}>Frequency</span>
      <div style={{ display: "flex", gap: 3 }}>
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            style={{
              width: 24,
              height: 8,
              borderRadius: 4,
              backgroundColor: i <= filled ? colors[filled - 1] : "#e0e0e0",
            }}
          />
        ))}
      </div>
      <span style={{ fontSize: 11, color: "#999" }}>{level}</span>
    </div>
  );
}

export default function CardBuilder({ candidates, onRegister, onDismiss, onRefreshImages, refreshing, registering }: CardBuilderProps) {
  const [meanings, setMeanings] = useState(candidates.meanings);
  const [examples, setExamples] = useState(candidates.examples);
  const [selectedMeaning, setSelectedMeaning] = useState(0);
  const [selectedExample, setSelectedExample] = useState(0);
  const [selectedImages, setSelectedImages] = useState<Set<number>>(new Set([0]));

  const toggleImage = (index: number, ctrlKey: boolean) => {
    setSelectedImages((prev) => {
      if (ctrlKey) {
        const next = new Set(prev);
        if (next.has(index)) {
          next.delete(index);
        } else {
          next.add(index);
        }
        return next;
      }
      // Single select: replace with clicked image
      return new Set([index]);
    });
  };

  const handleRegister = () => {
    const imageUrls = (candidates.image_urls || []).filter((_, i) => selectedImages.has(i));
    onRegister(
      candidates.word,
      meanings[selectedMeaning],
      examples[selectedExample],
      imageUrls,
    );
  };

  const updateMeaning = (index: number, value: string) => {
    const next = [...meanings];
    next[index] = value;
    setMeanings(next);
  };

  const updateExample = (index: number, value: string) => {
    const next = [...examples];
    next[index] = value;
    setExamples(next);
  };

  return (
    <div style={styles.container}>
      <div style={styles.word}>{candidates.word}</div>
      <div style={styles.meta}>
        <FrequencyGauge level={candidates.frequency} />
        <span style={{ ...styles.badge, backgroundColor: "#fff3e0", color: "#e65100" }}>
          Priority: {candidates.priority}
        </span>
      </div>
      <div style={styles.content}>
        <div style={styles.section}>
          <span style={styles.sectionLabel}>Core Image</span>
          <div style={styles.coreImage}>{candidates.core_image}</div>
        </div>

        <div style={styles.section}>
          <span style={styles.sectionLabel}>Usage</span>
          <div style={{ ...styles.coreImage, backgroundColor: "#f5f0ff", borderLeftColor: "#7c4dff" }}>
            {candidates.usage}
          </div>
        </div>

        <div style={styles.section}>
          <span style={styles.sectionLabel}>Meanings <span style={{ fontWeight: 400, color: "#aaa", fontSize: 11 }}>click to select, double-click to edit</span></span>
          {meanings.map((m, i) => (
            <EditableItem
              key={i}
              value={m}
              selected={i === selectedMeaning}
              onSelect={() => setSelectedMeaning(i)}
              onChange={(v) => updateMeaning(i, v)}
            />
          ))}
        </div>

        <div style={styles.section}>
          <span style={styles.sectionLabel}>Examples <span style={{ fontWeight: 400, color: "#aaa", fontSize: 11 }}>click to select, double-click to edit</span></span>
          {examples.map((e, i) => (
            <EditableItem
              key={i}
              value={e}
              selected={i === selectedExample}
              onSelect={() => setSelectedExample(i)}
              onChange={(v) => updateExample(i, v)}
            />
          ))}
        </div>

        {candidates.image_urls && candidates.image_urls.length > 0 && (
          <div style={styles.section}>
            <span style={styles.sectionLabel}>
              Images <span style={{ fontWeight: 400, color: "#aaa", fontSize: 11 }}>Ctrl+click for multi-select</span>
              <button
                onClick={onRefreshImages}
                disabled={refreshing}
                style={{ marginLeft: 8, padding: "2px 8px", fontSize: 11, border: "1px solid #ccc", borderRadius: 4, background: "none", color: refreshing ? "#ccc" : "#666", cursor: refreshing ? "default" : "pointer" }}
              >
                {refreshing ? "Loading..." : "Refresh"}
              </button>
            </span>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 8 }}>
              {candidates.image_urls.map((url, i) => (
                <div
                  key={i}
                  onClick={(e) => toggleImage(i, e.ctrlKey || e.metaKey)}
                  onDoubleClick={() => setSelectedImages(new Set())}
                  style={{
                    aspectRatio: "1",
                    borderRadius: 8,
                    overflow: "hidden",
                    border: selectedImages.has(i) ? "3px solid #4a90d9" : "3px solid transparent",
                    cursor: "pointer",
                  }}
                >
                  <img
                    src={url}
                    alt={candidates.word}
                    style={{ width: "100%", height: "100%", objectFit: "cover", display: "block" }}
                  />
                </div>
              ))}
            </div>
          </div>
        )}

        <div style={styles.section}>
          <span style={styles.sectionLabel}>Synonyms</span>
          {candidates.synonyms.map((s, i) => (
            <div key={i} style={styles.listItem}>
              <span style={{ fontWeight: 600 }}>{s.word}</span>
              <span style={{ color: "#888", marginLeft: 6 }}>— {s.difference}</span>
            </div>
          ))}
        </div>

      </div>
      <div style={styles.actions}>
        <button style={styles.closeBtn} onClick={onDismiss} disabled={registering}>
          Skip
        </button>
        <button
          style={{
            flex: 1, padding: "12px 16px", backgroundColor: "#4a90d9", color: "#fff",
            border: "none", borderRadius: 8, fontSize: 15, fontWeight: 600, cursor: "pointer",
          }}
          onClick={handleRegister}
          disabled={registering}
        >
          {registering ? "Registering..." : "Register to Anki"}
        </button>
      </div>
    </div>
  );
}
