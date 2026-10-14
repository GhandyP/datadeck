/* The DataDeck app shell. No framework, no build step, no other host. */
(function () {
  "use strict";

  var el = function (id) { return document.getElementById(id); };
  var titleInput = el("title");
  var audienceInput = el("audience");
  var dataInput = el("data");
  var generateButton = el("generate");
  var downloadButton = el("download");
  var notice = el("notice");
  var planRoot = el("plan");
  var planEmpty = el("plan-empty");
  var tableWrap = el("table-wrap");
  var table = el("source-table");
  var summary = el("summary");
  var codeBadge = el("code");
  var announce = el("announce");

  var state = { deck: null, selected: null };

  var MISSING_KEY_MESSAGE =
    "GOOGLE_API_KEY is not set, so no deck can be built. Set it in the environment " +
    "you started the server from, then reload this page.";

  function say(message) {
    notice.textContent = message;
    notice.hidden = !message;
  }

  function step(index, done) {
    ["step-1", "step-2", "step-3"].forEach(function (id, i) {
      var node = el(id);
      node.classList.toggle("is-active", i === index);
      node.classList.toggle("is-done", done && i < index);
    });
  }

  function node(tag, className, text) {
    var created = document.createElement(tag);
    if (className) { created.className = className; }
    if (text !== undefined && text !== null) { created.textContent = text; }
    return created;
  }

  /* ---- the parsed source, with a handle on every cell ---- */

  function cellId(row, column) { return "cell-" + row + "-" + column; }

  function renderTable(source) {
    table.textContent = "";
    var head = document.createElement("thead");
    var headRow = document.createElement("tr");
    source.columns.forEach(function (column, index) {
      var th = node("th", null, column);
      th.id = "cell-h-" + index;
      th.scope = "col";
      headRow.appendChild(th);
    });
    head.appendChild(headRow);
    table.appendChild(head);

    var body = document.createElement("tbody");
    source.rows.forEach(function (row, rowIndex) {
      var tr = document.createElement("tr");
      row.forEach(function (cell, columnIndex) {
        var td = node("td", /^[-+$€£¥]?[\d.,\s]+%?$/.test(cell.trim()) ? "numeric" : null, cell);
        td.id = cellId(rowIndex, columnIndex);
        tr.appendChild(td);
      });
      body.appendChild(tr);
    });
    table.appendChild(body);
    tableWrap.hidden = false;
  }

  function clearHighlights() {
    Array.prototype.forEach.call(table.querySelectorAll(".is-source"), function (cell) {
      cell.classList.remove("is-source");
    });
  }

  /* The signature move: light exactly the cells a claim rests on. */
  function highlight(evidenceIds) {
    clearHighlights();
    var facts = {};
    (state.deck.facts || []).forEach(function (fact) { facts[fact.id] = fact; });

    var described = [];
    evidenceIds.forEach(function (id) {
      var fact = facts[id];
      if (!fact) { return; }
      described.push(id + " " + fact.label + " " + fact.value);
      if (fact.row < 0) {
        var header = el("cell-h-" + fact.column);
        if (header) { header.classList.add("is-source"); }
        Array.prototype.forEach.call(table.querySelectorAll("tbody tr"), function (tr) {
          var cell = tr.children[fact.column];
          if (cell) { cell.classList.add("is-source"); }
        });
        return;
      }
      var cell = el(cellId(fact.row, fact.column));
      if (cell) { cell.classList.add("is-source"); }
    });
    return described;
  }

  function sourcesBlock(claim, facts) {
    var block = node("div", "sources");
    block.hidden = true;
    claim.evidenceIds.forEach(function (id) {
      var fact = facts[id] || { label: id, value: "" };
      var row = node("div", "sources__row");
      row.appendChild(node("span", "dd-code sources__id", id));
      row.appendChild(node("span", "sources__label", fact.label));
      row.appendChild(node("span", "sources__value", fact.value));
      block.appendChild(row);
    });
    return block;
  }

  function renderClaim(claim, facts) {
    var item = node("div", "claim");
    item.appendChild(node("p", "claim__text", claim.text));

    var meta = node("div", "claim__meta");
    var status = node(
      "span",
      "claim__status claim__status--" + claim.status,
      claim.status === "verified" ? "Verified" : "Rejected"
    );
    meta.appendChild(status);
    item.appendChild(meta);

    if (claim.status === "rejected") {
      item.classList.add("is-rejected");
      if (claim.reason) {
        item.appendChild(node("p", "claim__reason", claim.reason));
      }
      return item;
    }

    var sources = sourcesBlock(claim, facts);
    var button = node("button", "claim__prov", claim.evidenceIds.join(" "));
    button.type = "button";
    button.setAttribute("aria-pressed", "false");
    button.addEventListener("click", function () {
      var selected = state.selected === claim.id;
      state.selected = selected ? null : claim.id;
      button.setAttribute("aria-pressed", selected ? "false" : "true");
      sources.hidden = selected;
      item.classList.toggle("is-selected", !selected);

      Array.prototype.forEach.call(planRoot.querySelectorAll(".claim__prov"), function (other) {
        if (other !== button) { other.setAttribute("aria-pressed", "false"); }
      });
      Array.prototype.forEach.call(planRoot.querySelectorAll(".sources"), function (other) {
        if (other !== sources) { other.hidden = true; }
      });
      Array.prototype.forEach.call(planRoot.querySelectorAll(".claim"), function (other) {
        if (other !== item) { other.classList.remove("is-selected"); }
      });

      if (selected) {
        clearHighlights();
        announce.textContent = "Provenance hidden.";
      } else {
        announce.textContent = "Sources of " + claim.id + ": " + highlight(claim.evidenceIds).join("; ");
      }
    });
    meta.appendChild(button);
    item.appendChild(sources);
    return item;
  }

  function renderPlan(deck) {
    planRoot.textContent = "";
    var facts = {};
    (deck.facts || []).forEach(function (fact) { facts[fact.id] = fact; });

    var claims = 0;
    deck.slides.forEach(function (slide) {
      var block = node("div", "slide-block");

      var head = node("div", "slide-block__head");
      head.appendChild(node("span", "dd-code slide-block__number", "Slide " + slide.number));
      head.appendChild(node("h3", "slide-block__title", slide.title));
      block.appendChild(head);

      slide.claims.forEach(function (claim) {
        claims++;
        block.appendChild(renderClaim(claim, facts));
      });
      planRoot.appendChild(block);
    });

    planEmpty.hidden = claims > 0;
    if (claims === 0) {
      planEmpty.textContent = "No claim in this deck could be traced to the supplied data, so no slide was produced.";
      planEmpty.hidden = false;
    }
  }

  function renderSummary(deck) {
    summary.textContent = "";
    var rows = [
      ["Slides", String(deck.slides.length)],
      ["Claims shown", String(deck.slides.reduce(function (total, slide) {
        return total + slide.claims.filter(function (claim) { return claim.status === "verified"; }).length;
      }, 0))],
      ["Claims kept out", String(deck.rejectedCount)],
      ["Source values", String((deck.facts || []).length)]
    ];
    rows.forEach(function (pair) {
      var wrap = document.createElement("div");
      wrap.appendChild(node("dt", null, pair[0]));
      wrap.appendChild(node("dd", null, pair[1]));
      summary.appendChild(wrap);
    });
    summary.hidden = false;
  }

  /* ---- the flow ---- */

  function build() {
    var title = titleInput.value.trim();
    var raw = dataInput.value;

    if (!title) { say("Give the deck a title first."); titleInput.focus(); return; }
    if (!raw.trim()) { say("Paste the data the deck has to present."); dataInput.focus(); return; }

    say("");
    generateButton.disabled = true;
    generateButton.textContent = "Building…";

    fetch("/api/generate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: title, audience: audienceInput.value.trim() || "General", data: raw })
    })
      .then(function (response) {
        return response.json().catch(function () { return {}; }).then(function (body) {
          return { ok: response.ok, body: body };
        });
      })
      .then(function (result) {
        if (!result.ok) {
          say(result.body.error || "The deck could not be built.");
          step(0, false);
          return;
        }
        var deck = result.body;
        state.deck = deck;
        state.selected = null;

        codeBadge.textContent = deck.code;
        codeBadge.hidden = false;
        renderTable(deck.source);
        renderPlan(deck);
        renderSummary(deck);

        downloadButton.disabled = !deck.deckHTML;
        step(deck.deckHTML ? 2 : 1, Boolean(deck.deckHTML));
        if (!deck.deckHTML) {
          say("Every claim was rejected, so there is nothing to download. The reasons are in the plan.");
        }
      })
      .catch(function (error) {
        say("The server could not be reached: " + error.message);
      })
      .then(function () {
        generateButton.disabled = false;
        generateButton.textContent = "Build the deck";
      });
  }

  function download() {
    if (!state.deck || !state.deck.deckHTML) { return; }
    var blob = new Blob([state.deck.deckHTML], { type: "text/html" });
    var url = URL.createObjectURL(blob);
    var link = document.createElement("a");
    link.href = url;
    link.download = (state.deck.code || "datadeck") + ".html";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }

  generateButton.addEventListener("click", build);
  downloadButton.addEventListener("click", download);

  fetch("/api/health")
    .then(function (response) { return response.json(); })
    .then(function (health) {
      if (!health.apiKeyConfigured) {
        say(MISSING_KEY_MESSAGE);
        generateButton.disabled = true;
      }
    })
    .catch(function () {
      say("The server did not answer. Is it still running?");
    });

  step(0, false);
})();
