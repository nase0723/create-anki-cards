# Anki Card Creator

英語の単語をコピーするだけで、AIが意味・例文・画像を自動生成し、Ankiカードとして登録できるデスクトップアプリ。

> 「思考は残して、操作だけ削る」

## 処理フロー

```
単語をコピー（またはホットキー）
→ クリップボード検知
→ キャッシュ確認
→ AI が意味・例文・類義語を生成（GPT-4o-mini）
→ 画像候補を取得（Pixabay / Pexels）
→ ウィンドウに候補を表示
→ ユーザーが選択・編集
→ Ankiへ登録
```

## 技術構成

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.23 |
| デスクトップ | Wails v2 |
| フロントエンド | React 18 + TypeScript + Vite |
| AI | OpenAI API (GPT-4o-mini) |
| 画像検索 | Pixabay API / Pexels API |
| Anki連携 | AnkiConnect |
| キャッシュ | ファイルベースJSON |

## セットアップ

### 前提条件

- Go 1.23+
- Node.js
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- Anki + [AnkiConnect](https://ankiweb.net/shared/info/2055492159) アドオン

### 設定

`config.example.json` をコピーして `config.json` を作成し、APIキーを設定する。

```bash
cp config.example.json config.json
```

```json
{
  "openai_api_key": "sk-...",
  "ankiconnect_url": "http://localhost:8765",
  "deck_name": "English Vocabulary",
  "note_type": "EnglishVocab",
  "field_mapping": {
    "word": "Word",
    "meaning": "Meaning",
    "example": "Example",
    "image": "Image"
  },
  "trigger_mode": "polling",
  "poll_interval_ms": 500,
  "pixabay_api_key": "",
  "pexels_api_key": ""
}
```

- `trigger_mode`: `"polling"`（クリップボード監視）または `"hotkey"`（Ctrl+Shift+W）
- `pixabay_api_key` / `pexels_api_key`: 画像検索を使う場合に設定（どちらか片方でもOK）

### ビルド・実行

```bash
# 開発モード
wails dev

# 本番ビルド
wails build
```

## プロジェクト構成

```
├── main.go                  # エントリーポイント（多重起動防止あり）
├── app.go                   # Wailsアプリ本体・バックエンドAPI
├── config.example.json      # 設定ファイルのテンプレート
├── internal/
│   ├── ai/                  # OpenAI API クライアント
│   ├── anki/                # AnkiConnect クライアント
│   ├── cache/               # ファイルベースキャッシュ
│   ├── clipboard/           # クリップボード監視 / ホットキートリガー
│   ├── config/              # 設定読み込み
│   └── image/               # Pixabay / Pexels 画像検索
└── frontend/
    └── src/
        ├── App.tsx           # メインコンポーネント
        └── components/
            ├── CardBuilder.tsx   # カード編集UI
            └── EditableItem.tsx  # 選択・編集可能な項目
```