# Frontend Build Instructions / 前端构建说明 / フロントエンド構築手順

## English

### Frontend Repository

- **Frontend project repository**: https://github.com/komari-monitor/komari-web

### Build Requirements

1. Clone the frontend repository and build the static files
2. Pack the generated `dist` directory into `web/public/defaultTheme/dist.tar.zst` with `go run ./cmd/pack-frontend` (it also copies `komari-theme.json`)
3. The packer verifies that `index.html` and the ES module import graph only reference files present in the archive, and fails the build otherwise
4. `scripts/build-default-theme.sh` (or `.ps1` on Windows) automates the whole clone + build + pack flow

### Important Note

⚠️ **The projects under Akizon77's personal repository are no longer maintained. Please use the projects under the organization (komari-monitor).**

---

## 中文

### 前端项目仓库

- **前端项目地址**: https://github.com/komari-monitor/komari-web

### 构建要求

1. 克隆前端仓库并构建静态文件
2. 用 `go run ./cmd/pack-frontend` 将生成的 `dist` 目录打包为 `web/public/defaultTheme/dist.tar.zst`（同时会复制 `komari-theme.json`）
3. 打包工具会校验 `index.html` 与 ES module import 图只引用归档内存在的文件，不一致则直接失败
4. `scripts/build-default-theme.sh`（Windows 为 `.ps1`）可一键完成 克隆 + 构建 + 打包

### 重要提醒

⚠️ **Akizon77 个人仓库的项目已经不再使用，请使用组织（komari-monitor）下的项目。**

---

## 日本語

### フロントエンドプロジェクトリポジトリ

- **フロントエンドプロジェクトアドレス**: https://github.com/komari-monitor/komari-web

### ビルド要件

1. フロントエンドリポジトリをクローンして静的ファイルをビルドする
2. 生成された `dist` ディレクトリを `go run ./cmd/pack-frontend` で `web/public/defaultTheme/dist.tar.zst` にパックする（`komari-theme.json` もコピーされる）
3. パッカーは `index.html` と ES module の import グラフが参照するファイルがすべてアーカイブ内に存在するか検証し、不一致なら失敗する
4. `scripts/build-default-theme.sh`（Windows は `.ps1`）でクローン + ビルド + パックを一括実行できる

### 重要な注意事項

⚠️ **Akizon77 の個人リポジトリのプロジェクトは使用されなくなりました。組織（komari-monitor）下のプロジェクトを使用してください。**

---

## Quick Setup / 快速设置 / クイックセットアップ

```bash
# One command from the backend repo root / 在后端仓库根目录一条命令完成 / バックエンドリポジトリのルートで1コマンド
bash scripts/build-default-theme.sh
# Windows (PowerShell): pwsh scripts/build-default-theme.ps1

# Or do it manually / 或手动执行 / または手動で実行
git clone https://github.com/komari-monitor/komari-web
cd komari-web
npm install
npm run build
cd ..

# Pack + verify into the embed archive / 打包并校验 / パックして検証
go run ./cmd/pack-frontend \
  -dist komari-web/dist \
  -out web/public/defaultTheme/dist.tar.zst \
  -theme komari-web/komari-theme.json

# The same check runs in CI / 同样的校验在 CI 中执行 / 同じ検証が CI で実行される
go test ./web/public/ -run TestEmbeddedFrontendIsSelfConsistent -count=1
```
