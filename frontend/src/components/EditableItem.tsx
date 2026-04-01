import { useState, useRef, useEffect, CSSProperties } from "react";

interface EditableItemProps {
  value: string;
  onChange: (value: string) => void;
  selected?: boolean;
  onSelect?: () => void;
}

const styles: Record<string, CSSProperties> = {
  item: {
    padding: "6px 12px",
    borderRadius: 6,
    border: "1px solid #e0e0e0",
    marginBottom: 4,
    fontSize: 14,
    lineHeight: 1.5,
    cursor: "pointer",
  },
  itemSelected: {
    padding: "6px 12px",
    borderRadius: 6,
    border: "1px solid #4a90d9",
    backgroundColor: "#eef4fd",
    marginBottom: 4,
    fontSize: 14,
    lineHeight: 1.5,
    cursor: "pointer",
  },
  input: {
    width: "100%",
    padding: "6px 12px",
    borderRadius: 6,
    border: "1px solid #4a90d9",
    marginBottom: 4,
    fontSize: 14,
    lineHeight: 1.5,
    outline: "none",
    boxSizing: "border-box" as const,
  },
};

export default function EditableItem({ value, onChange, selected, onSelect }: EditableItemProps) {
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  const handleClick = () => {
    if (onSelect) onSelect();
  };

  const handleDoubleClick = () => {
    setEditValue(value);
    setEditing(true);
  };

  const handleBlur = () => {
    setEditing(false);
    if (editValue.trim() && editValue !== value) {
      onChange(editValue.trim());
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      (e.target as HTMLInputElement).blur();
    } else if (e.key === "Escape") {
      setEditValue(value);
      setEditing(false);
    }
  };

  if (editing) {
    return (
      <input
        ref={inputRef}
        style={styles.input}
        value={editValue}
        onChange={(e) => setEditValue(e.target.value)}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
    );
  }

  return (
    <div
      style={selected ? styles.itemSelected : styles.item}
      onClick={handleClick}
      onDoubleClick={handleDoubleClick}
      title="Double-click to edit"
    >
      {value}
    </div>
  );
}
