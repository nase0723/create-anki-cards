import { CSSProperties } from "react";

interface SelectorProps {
  label: string;
  items: string[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

const styles: Record<string, CSSProperties> = {
  container: {
    marginBottom: 16,
  },
  label: {
    fontSize: 12,
    fontWeight: 600,
    color: "#666",
    textTransform: "uppercase" as const,
    letterSpacing: 0.5,
    marginBottom: 8,
    display: "block",
  },
  item: {
    padding: "8px 12px",
    borderRadius: 6,
    cursor: "pointer",
    border: "1px solid #e0e0e0",
    marginBottom: 4,
    transition: "all 0.15s",
    fontSize: 14,
    lineHeight: 1.5,
  },
  itemSelected: {
    padding: "8px 12px",
    borderRadius: 6,
    cursor: "pointer",
    border: "1px solid #4a90d9",
    backgroundColor: "#eef4fd",
    marginBottom: 4,
    fontSize: 14,
    lineHeight: 1.5,
  },
};

export default function Selector({ label, items, selectedIndex, onSelect }: SelectorProps) {
  return (
    <div style={styles.container}>
      <span style={styles.label}>{label}</span>
      {items.map((item, i) => (
        <div
          key={i}
          style={i === selectedIndex ? styles.itemSelected : styles.item}
          onClick={() => onSelect(i)}
        >
          {item}
        </div>
      ))}
    </div>
  );
}
