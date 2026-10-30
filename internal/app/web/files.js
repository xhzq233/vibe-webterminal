(() => {
  "use strict";
  let browserPanel, pathInput, listing, status, download, preview, parentButton;
  let currentPath = "", parentPath = "", imageURL = "", requestNumber = 0;
  const lastDirectoryKey = "herdr-tty-last-directory";
  let lastDirectory = "";
  try { lastDirectory = window.localStorage.getItem(lastDirectoryKey) || ""; } catch { /* Keep memory for this page if storage is unavailable. */ }
  const fileURL = (path) => "/_herdr/download?" + new URLSearchParams({ path });
  function sizeText(size) {
    if (size < 1024) return size + " B";
    if (size < 1024 * 1024) return (size / 1024).toFixed(1) + " KB";
    return (size / 1024 / 1024).toFixed(1) + " MB";
  }
  function clearPreview() {
    preview.replaceChildren();
    if (imageURL) URL.revokeObjectURL(imageURL);
    imageURL = "";
  }
  async function openTabDirectory() {
    const number = ++requestNumber;
    const inputAtStart = pathInput.value;
    status.textContent = "正在读取当前 Tab 的目录…";
    listing.replaceChildren();
    clearPreview();
    download.hidden = true;
    try {
      const response = await fetch("/_herdr/cwd", { cache: "no-store" });
      if (number !== requestNumber || browserPanel.hidden) return;
      if (response.redirected) throw new Error("登录已过期，请刷新网页重新登录");
      if (!response.ok) throw new Error((await response.text()).trim() || "无法读取当前目录");
      const data = await response.json();
      if (number !== requestNumber || browserPanel.hidden) return;
      if (pathInput.value !== inputAtStart) {
        status.textContent = "输入路径后点打开";
        return;
      }
      pathInput.value = data.path;
      void openPath(data.path);
    } catch (error) {
      if (number === requestNumber && !browserPanel.hidden) status.textContent = error.message || "无法读取当前目录，请手动输入路径";
    }
  }
  async function openPath(path = "") {
    const number = ++requestNumber;
    const inputAtStart = pathInput.value;
    status.textContent = "读取中…";
    listing.replaceChildren();
    clearPreview();
    download.hidden = true;
    try {
      const response = await fetch("/_herdr/files?" + new URLSearchParams({ path }), { cache: "no-store" });
      if (number !== requestNumber || browserPanel.hidden) return;
      if (response.redirected) throw new Error("登录已过期，请刷新网页重新登录");
      if (!response.ok) throw new Error((await response.text()).trim() || "读取失败");
      const data = await response.json();
      if (number !== requestNumber || browserPanel.hidden) return;
      currentPath = data.path;
      parentPath = data.parent;
      lastDirectory = data.directory ? data.path : data.parent;
      try { window.localStorage.setItem(lastDirectoryKey, lastDirectory); } catch { /* In-page memory still works. */ }
      if (pathInput.value === inputAtStart) pathInput.value = currentPath;
      parentButton.disabled = currentPath === parentPath;
      if (data.directory) {
        status.textContent = data.entries.length ? data.entries.length + " 项" : "空目录";
        for (const file of data.entries) {
          const row = document.createElement("button");
          row.type = "button";
          row.className = file.directory ? "file-entry file-directory" : "file-entry";
          const name = document.createElement("span");
          name.className = "file-name";
          name.textContent = file.name;
          const detail = document.createElement("span");
          detail.className = "file-detail";
          detail.textContent = file.directory ? "目录" : sizeText(file.size);
          row.append(name, detail);
          row.addEventListener("click", () => { void openPath(file.path); });
          listing.appendChild(row);
        }
      } else {
        download.href = fileURL(currentPath);
        download.download = data.name;
        download.hidden = false;
        status.textContent = sizeText(data.size);
        if (data.size <= 20 * 1024 * 1024 && /^image\/(png|jpeg|gif|webp|bmp|avif)$/.test(data.contentType)) {
          const imageResponse = await fetch(fileURL(currentPath), { cache: "no-store" });
          if (imageResponse.redirected || !imageResponse.ok) throw new Error("图片读取失败，请检查登录状态");
          const blob = await imageResponse.blob();
          if (number !== requestNumber || browserPanel.hidden) return;
          imageURL = URL.createObjectURL(new Blob([blob], { type: data.contentType }));
          const image = document.createElement("img");
          image.src = imageURL;
          image.alt = data.name;
          preview.appendChild(image);
        } else if (Object.hasOwn(data, "text")) {
          const text = document.createElement("pre");
          text.textContent = data.text;
          preview.appendChild(text);
          if (data.truncated) status.textContent += " · 仅预览前 256 KB，下载可获取完整文件";
          else if (!data.text) status.textContent += " · 空文件";
        } else status.textContent += " · 此类型请下载后查看";
      }
    } catch (error) {
      if (number === requestNumber && !browserPanel.hidden) status.textContent = error.message || "无法连接文件服务";
    }
  }
  function createBrowser() {
    browserPanel = document.createElement("section");
    browserPanel.hidden = true;
    browserPanel.className = "herdr-tty-files";
    browserPanel.setAttribute("aria-label", "远端文件");
    browserPanel.innerHTML = '<header><strong>远端文件</strong><button type="button" data-file-action="close" aria-label="关闭文件浏览器">✕</button></header>' +
      '<form><input aria-label="文件路径" placeholder="输入路径，例如 ~/ 或 /home/…" autocomplete="off" autocapitalize="off" spellcheck="false"><button type="submit">打开</button></form>' +
      '<nav><button type="button" data-file-action="cwd">当前 Tab</button><button type="button" data-file-action="home">主目录</button><button type="button" data-file-action="parent">上一级</button><button type="button" data-file-action="refresh">刷新</button><a hidden data-file-action="download">下载文件</a></nav>' +
      '<p role="status"></p><section class="file-list" aria-label="目录内容"></section><section class="file-preview" aria-label="文件预览"></section>';
    document.getElementById("panel-content").appendChild(browserPanel);
    pathInput = browserPanel.querySelector("input");
    listing = browserPanel.querySelector(".file-list");
    preview = browserPanel.querySelector(".file-preview");
    status = browserPanel.querySelector('[role="status"]');
    download = browserPanel.querySelector('[data-file-action="download"]');
    parentButton = browserPanel.querySelector('[data-file-action="parent"]');
    browserPanel.querySelector("form").addEventListener("submit", (event) => {
      event.preventDefault();
      pathInput.blur();
      void openPath(pathInput.value.trim());
    });
    browserPanel.querySelector('[data-file-action="close"]').addEventListener("click", closeBrowser);
    browserPanel.querySelector('[data-file-action="home"]').addEventListener("click", () => { void openPath("~"); });
    browserPanel.querySelector('[data-file-action="cwd"]').addEventListener("click", () => { void openTabDirectory(); });
    parentButton.addEventListener("click", () => { void openPath(parentPath); });
    browserPanel.querySelector('[data-file-action="refresh"]').addEventListener("click", () => { void openPath(currentPath); });
    browserPanel.addEventListener("keydown", (event) => event.stopPropagation());
    browserPanel.addEventListener("keyup", (event) => event.stopPropagation());
    browserPanel.addEventListener("keypress", (event) => event.stopPropagation());

  }
  function closeBrowser() {
    browserPanel.hidden = true;
    ++requestNumber;
    clearPreview();
    document.getElementById("panel-files-button").setAttribute("aria-expanded", "false");
    window.dispatchEvent(new Event("herdr-panel-resize"));
  }
  window.addEventListener("herdr-open-files", () => {
    if (!browserPanel) createBrowser();
    if (!browserPanel.hidden) { closeBrowser(); return; }
    browserPanel.hidden = false;
    document.getElementById("panel-files-button").setAttribute("aria-expanded", "true");
    window.dispatchEvent(new Event("herdr-panel-resize"));
    void openPath(lastDirectory);
  });
})();
