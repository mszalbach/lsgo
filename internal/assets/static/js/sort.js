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
  const sortRows = (button, order) => {
    const key = button.dataset.sortKey;
    const ascending = order === "ascending";

    sortButtons.forEach((sortButton) => {
      sortButton.setAttribute(
        "aria-sort",
        sortButton === button ? order : "none",
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
      const order = button.getAttribute("aria-sort") === "ascending" ? "descending" : "ascending";
      sortRows(button, order);
    });
  });

  const defaultButton = document.querySelector('.sort-controls button[data-sort-key="name"]');
  if (defaultButton) sortRows(defaultButton, "ascending");
})();