// Build a self-contained HTML user manual (images inlined as data URIs).
// Run from frontend/: node playwright/build-manual-html.mjs
import fs from "fs";
import path from "path";

const md = fs.readFileSync(path.resolve("../docs/user-manual.md"), "utf8");
const imgDir = path.resolve("../docs/img");
const outFile = path.resolve("../docs/user-manual.html");

const esc = (s) =>
  s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
const inline = (s) =>
  esc(s)
    .replace(/`([^`]+)`/g, "<code>$1</code>")
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");

function dataUri(rel) {
  const p = path.join(imgDir, rel.replace(/^img\//, ""));
  if (!fs.existsSync(p)) return null;
  const b64 = fs.readFileSync(p).toString("base64");
  return `data:image/png;base64,${b64}`;
}

const lines = md.split("\n");
let html = "";
let inList = false;
const closeList = () => {
  if (inList) {
    html += "</ul>\n";
    inList = false;
  }
};

for (let raw of lines) {
  const line = raw.replace(/\r$/, "");
  const img = line.match(/^!\[([^\]]*)\]\(([^)]+)\)/);
  if (img) {
    closeList();
    const uri = dataUri(img[2]);
    html += uri
      ? `<figure><img alt="${esc(img[1])}" src="${uri}"><figcaption>${esc(img[1])}</figcaption></figure>\n`
      : `<p><em>[missing image: ${esc(img[2])}]</em></p>\n`;
    continue;
  }
  if (/^---\s*$/.test(line)) {
    closeList();
    html += "<hr>\n";
    continue;
  }
  if (/^###\s+/.test(line)) {
    closeList();
    html += `<h3>${inline(line.replace(/^###\s+/, ""))}</h3>\n`;
    continue;
  }
  if (/^##\s+/.test(line)) {
    closeList();
    html += `<h2>${inline(line.replace(/^##\s+/, ""))}</h2>\n`;
    continue;
  }
  if (/^#\s+/.test(line)) {
    closeList();
    html += `<h1>${inline(line.replace(/^#\s+/, ""))}</h1>\n`;
    continue;
  }
  if (/^>\s?/.test(line)) {
    closeList();
    html += `<blockquote>${inline(line.replace(/^>\s?/, ""))}</blockquote>\n`;
    continue;
  }
  if (/^[-*]\s+/.test(line)) {
    if (!inList) {
      html += "<ul>\n";
      inList = true;
    }
    html += `<li>${inline(line.replace(/^[-*]\s+/, ""))}</li>\n`;
    continue;
  }
  if (line.trim() === "") {
    closeList();
    continue;
  }
  closeList();
  html += `<p>${inline(line)}</p>\n`;
}
closeList();

const doc = `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>OpenELIS Global 2 — User Manual</title>
<style>
 :root{--fg:#1a1a1a;--muted:#5a6b7b;--bd:#dfe4ea;--accent:#15497a;--bg:#fff;--code:#eef2f6}
 @media (prefers-color-scheme:dark){:root{--fg:#e6e6e6;--muted:#9fb0c0;--bd:#2c3440;--accent:#6ab0ff;--bg:#0f1319;--code:#1a2029}}
 *{box-sizing:border-box} body{margin:0;background:var(--bg);color:var(--fg);
   font:16px/1.6 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif}
 .wrap{max-width:980px;margin:0 auto;padding:32px 20px 80px}
 h1{font-size:2rem;border-bottom:3px solid var(--accent);padding-bottom:.3em;margin-top:1.6em}
 h2{font-size:1.5rem;color:var(--accent);margin-top:2em;border-bottom:1px solid var(--bd);padding-bottom:.2em}
 h3{font-size:1.15rem;margin-top:1.6em}
 blockquote{margin:1em 0;padding:.6em 1em;border-left:4px solid var(--accent);background:var(--code);color:var(--muted);border-radius:4px}
 code{background:var(--code);padding:.1em .4em;border-radius:4px;font:0.88em ui-monospace,Consolas,monospace}
 figure{margin:1.2em 0}
 figure img{width:100%;max-width:100%;border:1px solid var(--bd);border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,.10)}
 figcaption{color:var(--muted);font-size:.85rem;margin-top:.4em;text-align:center}
 hr{border:none;border-top:1px solid var(--bd);margin:2.4em 0}
 a{color:var(--accent)} ul{padding-left:1.3em}
</style></head><body><div class="wrap">
${html}
</div></body></html>`;

fs.writeFileSync(outFile, doc);
const kb = Math.round(fs.statSync(outFile).size / 1024);
console.log(`wrote ${outFile} (${kb} KB)`);
