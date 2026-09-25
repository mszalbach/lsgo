(() => {
  const table = document.querySelector("table");
  const masterCheckbox = table?.querySelector('input[type="checkbox"][name="master"]');
  const itemCheckboxes = table?.querySelectorAll(
    'tbody input[type="checkbox"][name="paths"]',
  );
  const selectionBar = document.querySelector("#selection-bar");
  const selectedCountLabel = document.querySelector("#selected-count");
  const deselectButton = document.querySelector("#deselect-all");

  if (!masterCheckbox || !itemCheckboxes?.length) return;

  const updateMasterCheckbox = () => {
    const selectedCount = Array.from(itemCheckboxes).filter(
      (checkbox) => checkbox.checked,
    ).length;

    masterCheckbox.checked = selectedCount === itemCheckboxes.length;
    masterCheckbox.indeterminate = selectedCount > 0 && selectedCount < itemCheckboxes.length;
    const isEmpty = selectedCount === 0;
    selectionBar.setAttribute("aria-hidden", isEmpty);
    selectionBar.inert = isEmpty;
    selectedCountLabel.textContent = `${selectedCount} selected`;
  };

  masterCheckbox.addEventListener("change", () => {
    itemCheckboxes.forEach((checkbox) => {
      checkbox.checked = masterCheckbox.checked;
    });
    updateMasterCheckbox();
  });

  itemCheckboxes.forEach((checkbox) => {
    checkbox.addEventListener("change", updateMasterCheckbox);
  });

  deselectButton.addEventListener("click", () => {
    itemCheckboxes.forEach((checkbox) => {
      checkbox.checked = false;
    });
    updateMasterCheckbox();
  });

  updateMasterCheckbox();
})();