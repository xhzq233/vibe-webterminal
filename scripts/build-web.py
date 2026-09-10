#!/usr/bin/env python3
from pathlib import Path

web = Path(__file__).resolve().parents[1] / "web"
base = (web / "index-original.html").read_text().replace(
    "<head>", '<head><meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">'
)
addon = "<script>\n" + (web / "interact.min.js").read_text() + "\n</script>\n" + (web / "touch-panel.html").read_text()
(web / "index.html").write_text(base.replace("</body>", addon + "</body>"))
print("Updated web/index.html")
