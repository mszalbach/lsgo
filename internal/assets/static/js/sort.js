(() => {
  const collator = new Intl.Collator(undefined, {
    numeric: true,
    sensitivity: "base",
  });

  const compareValues = (key, left, right) => {
    if (key === "name") return collator.compare(left, right);
    return Number(left) - Number(right);
  };

  const table = document.querySelector("#folder-contents");
  const body = table?.querySelector("tbody");
  if (!table || !body) return;

  const sortButtons = document.querySelectorAll(".sort-controls button[data-sort-key]");
  const sortHeaders = table.querySelectorAll("thead th[data-sort-key]");
  const storageKey = "lsgo-folder-sort";
  const sortRows = (button, order) => {
    const key = button.dataset.sortKey;
    const ascending = order === "ascending";

    sortButtons.forEach((sortButton) => {
      const isActive = sortButton === button;
      sortButton.setAttribute("aria-pressed", String(isActive));
      if (isActive) {
        sortButton.setAttribute(
          "aria-label",
          `${sortButton.textContent.trim()}, sorted ${order}`,
        );
        sortButton.dataset.sortOrder = order;
      } else {
        sortButton.removeAttribute("aria-label");
        delete sortButton.dataset.sortOrder;
      }
    });

    sortHeaders.forEach((header) => {
      header.setAttribute(
        "aria-sort",
        header.dataset.sortKey === key ? order : "none",
      );
    });

    const rows = Array.from(body.querySelectorAll("tr")).sort((firstRow, secondRow) => {
      const groupDiff = Number(firstRow.dataset.sortGroup) - Number(secondRow.dataset.sortGroup);
      if (groupDiff !== 0) return groupDiff;

      const firstCell = firstRow.querySelector(`[data-sort-column="${key}"]`);
      const secondCell = secondRow.querySelector(`[data-sort-column="${key}"]`);
      const firstValue = firstCell?.dataset.sortValue ?? "";
      const secondValue = secondCell?.dataset.sortValue ?? "";

      const comparison = compareValues(key, firstValue, secondValue);
      return ascending ? comparison : -comparison;
    });

    rows.forEach((row) => body.appendChild(row));
  };

  sortButtons.forEach((button) => {
    button.addEventListener("click", () => {
      const order = button.dataset.sortOrder === "ascending" ? "descending" : "ascending";
      sortRows(button, order);
      try {
        window.localStorage.setItem(storageKey, `${button.dataset.sortKey}:${order}`);
      } catch (error) {
        console.warn("Unable to save the folder sort preference.", error);
      }
    });
  });

  let savedSort = "";
  try {
    savedSort = window.localStorage.getItem(storageKey) ?? "";
  } catch (error) {
    console.warn("Unable to read the saved folder sort preference.", error);
  }

  const [savedKey, savedOrder] = savedSort.split(":");
  const savedButton = Array.from(sortButtons).find(
    (button) => button.dataset.sortKey === savedKey,
  );
  const initialButton = savedButton ?? document.querySelector(
    '.sort-controls button[data-sort-key="name"]',
  );
  const initialOrder =
    savedButton && (savedOrder === "ascending" || savedOrder === "descending")
      ? savedOrder
      : "ascending";
  if (initialButton) sortRows(initialButton, initialOrder);
})();