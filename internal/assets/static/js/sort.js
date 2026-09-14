(() => {
  const collator = new Intl.Collator(undefined, {
    numeric: true,
    sensitivity: "base",
  });

  const compareValues = (key, left, right) => {
    if (key === "name") return collator.compare(left, right);
    return Number(left) - Number(right);
  };

  const resetHeaders = (headers) => {
    headers.forEach((header) => {
      header.dataset.order = "none";
      header.setAttribute("aria-sort", "none");
    });
  };

  document.querySelectorAll("table thead th[data-sort-key]").forEach((header) => {
    const table = header.closest("table");
    const body = table?.querySelector("tbody");
    if (!table || !body) return;

    const sortRows = () => {
      const headers = table.querySelectorAll("thead th[data-sort-key]");
      const key = header.dataset.sortKey;
      const ascending = header.dataset.order !== "ascending";

      resetHeaders(headers);

      header.dataset.order = ascending ? "ascending" : "descending";
      header.setAttribute("aria-sort", header.dataset.order);

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

    header.addEventListener("click", sortRows);
    header.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        sortRows();
      }
    });
  });

  const defaultHeader = document.querySelector('table thead th[data-sort-key="name"]');
  defaultHeader?.click();
})();