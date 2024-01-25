const Controller = {
  currentPage: 0,
  
  pageSize: 20,

  search: (ev, loadMore = false) => {
    if (ev) ev.preventDefault();
    if (!loadMore) {
      Controller.currentPage = 0; // Reset to first page for new searches
      Controller.clearTable(); // Clear existing results if any
    }

    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}&page=${Controller.currentPage}&pageSize=${Controller.pageSize}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results, loadMore);
        if (results.length > 0) {
          Controller.currentPage++;
        }
      });
    });
  },

  updateTable: (results, append = false) => {
    const table = document.getElementById("table-body");
    const rows = [];
    for (let result of results) {
      rows.push(`<tr><td>${result}</td></tr>`);
    }
    if (append) {
      table.innerHTML += rows.join('');
    } else {
      table.innerHTML = rows.join('');
    }
  },
  
  clearTable: () => {
    document.getElementById("table-body").innerHTML = '';
  }
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);

const loadMoreButton = document.getElementById("load-more");
loadMoreButton.addEventListener("click", (ev) => Controller.search(ev, true));
