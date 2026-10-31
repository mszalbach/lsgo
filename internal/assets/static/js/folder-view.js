(() => {
  const table = document.querySelector("#folder-contents");
  const controls = document.querySelector("#folder-view-controls");
  const buttons = controls?.querySelectorAll("[data-folder-view]");
  if (!table || !controls || !buttons?.length) return;

  const storageKey = "lsgo-folder-view";
  let view = "list";
  try {
    const savedView = window.localStorage.getItem(storageKey);
    if (savedView === "list" || savedView === "grid") view = savedView;
  } catch (error) {
    console.warn("Unable to read the saved folder view preference.", error);
  }

  const setView = (nextView) => {
    if (nextView !== "list" && nextView !== "grid") return;
    table.dataset.view = nextView;
    controls.dataset.view = nextView;
    buttons.forEach((button) => {
      button.setAttribute("aria-pressed", String(button.dataset.folderView === nextView));
    });

    try {
      window.localStorage.setItem(storageKey, nextView);
    } catch (error) {
      console.warn("Unable to save the folder view preference.", error);
    }
  };

  buttons.forEach((button) => {
    button.addEventListener("click", () => setView(button.dataset.folderView));
  });

  controls.hidden = false;
  setView(view);
})();
