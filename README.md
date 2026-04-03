# Anki Card Creator

英語の単語をコピーするだけで、AIが意味・例文・画像を自動生成し、Ankiカードとして登録できるデスクトップアプリ。

> 「思考は残して、操作だけ削る」

## 処理フロー

```
単語をコピー（またはCtrl+B）
→ 重複チェック（登録済みならAnkiブラウザを表示）
→ キャッシュ確認
→ AI が意味・例文・類義語・コアイメージ・使用場面を生成（Gemini 2.5 Flash Lite）
→ 画像候補を取得（Pixabay / Pexels、並列検索）
→ ウィンドウに候補を表示
→ ユーザーが選択・編集
→ Ankiへ登録（画像含む）
```

## 機能

- クリップボード監視（ポーリング）+ Ctrl+B ホットキー
- AI による意味・例文・コアイメージ・類義語（違い付き）・使用頻度・優先度・使用場面の生成
- Pixabay / Pexels からの画像検索（2プロバイダー並列）
- 画像の Refresh（AIが新しいキーワードで再検索）
- 意味・例文の選択・ダブルクリック編集
- 画像の単一選択 / Ctrl+クリックで複数選択
- Anki 重複チェック（登録済みならAnkiブラウザを直接表示）
- Anki への画像付きカード登録
- Esc キーでウィンドウ非表示
- システムトレイ常駐（右クリック→Quit で終了）

## 技術構成

| レイヤー | 技術 |
|---|---|
| バックエンド | Go |
| デスクトップ | Wails v2 |
| フロントエンド | React + TypeScript + Vite |
| AI | Gemini API（デフォルト: gemini-2.5-flash-lite、設定で変更可能） |
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
  "gemini_api_key": "",
  "gemini_model": "gemini-2.5-flash-lite",
  "ankiconnect_url": "http://localhost:8765",
  "deck_name": "English Vocabulary",
  "note_type": "EnglishVocab",
  "field_mapping": {
    "word": "Word",
    "meaning": "Meaning",
    "example": "Example",
    "image": "Image"
  },
  "poll_interval_ms": 500,
  "polling_enabled": true,
  "pixabay_api_key": "",
  "pexels_api_key": ""
}
```

- `gemini_api_key`: [Google AI Studio](https://aistudio.google.com/) で取得
- `gemini_model`: 使用するモデル名（設定で変更可能）
- `polling_enabled`: `false` にするとクリップボード監視を無効化（Ctrl+B のみで動作）
- `pixabay_api_key` / `pexels_api_key`: 画像検索を使う場合に設定（どちらか片方でもOK）

### ビルド・実行

```bash
# 開発モード
wails dev

# 本番ビルド
wails build
```

本番ビルド後は `build/bin/create-anki-cards.exe` と同じフォルダに `config.json` を配置して実行。

## 操作方法

| 操作 | 動作 |
|---|---|
| 単語をコピー | ウィンドウが自動表示（ポーリング有効時） |
| Ctrl+B | クリップボードの単語でウィンドウ表示 |
| 意味・例文をクリック | 選択 |
| 意味・例文をダブルクリック | 編集モード |
| 画像をクリック | 選択（1枚） |
| 画像を Ctrl+クリック | 複数選択 |
| 画像をダブルクリック | 選択解除 |
| Refresh ボタン | AIが新しいキーワードで画像を再検索 |
| Register to Anki | 選択した内容でAnkiに登録 |
| Skip / Esc | ウィンドウを非表示 |
| システムトレイ → Quit | アプリ終了 |

## プロジェクト構成

```
├── main.go                  # エントリーポイント
├── app.go                   # Wailsアプリ本体・バックエンドAPI
├── config.example.json      # 設定ファイルのテンプレート
├── internal/
│   ├── ai/                  # Gemini API クライアント
│   ├── anki/                # AnkiConnect クライアント
│   ├── cache/               # ファイルベースキャッシュ
│   ├── clipboard/           # クリップボード監視
│   ├── config/              # 設定読み込み
│   └── image/               # Pixabay / Pexels 画像検索
└── frontend/
    └── src/
        ├── App.tsx           # メインコンポーネント
        └── components/
            ├── CardBuilder.tsx   # カード編集UI
            ├── EditableItem.tsx  # 選択・編集可能な項目
            └── Selector.tsx     # 汎用選択コンポーネント
```
