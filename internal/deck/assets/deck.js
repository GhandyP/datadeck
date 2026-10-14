/* The exported deck: one self-contained file, no framework, no network. */
(function () {
  "use strict";

  var deck = window.__DATADECK__;
  if (!deck) {
    return;
  }

  var slidesRoot = document.getElementById("slides");
  var indexRoot = document.getElementById("index");
  var counter = document.getElementById("counter");
  var previous = document.getElementById("prev");
  var next = document.getElementById("next");

  document.getElementById("deck-code").textContent = deck.code;
  document.getElementById("deck-title").textContent = deck.title;
  document.getElementById("deck-draft").textContent = draftLabel(deck);
  document.title = deck.title + " — DataDeck draft";

  var factsById = {};
  (deck.facts || []).forEach(function (fact) {
    factsById[fact.id] = fact;
  });

  var slides = (deck.slides || []).map(function (slide) {
    return {
      number: slide.number,
      title: slide.title,
      notes: slide.notes,
      claims: (slide.claims || []).filter(function (claim) {
        return claim.status === "verified";
      })
    };
  });

  var current = 0;

  if (slides.length === 0) {
    var empty = document.createElement("p");
    empty.className = "fallback";
    empty.textContent =
      "No claim in this deck could be traced to the supplied data, so there is nothing to show.";
    slidesRoot.appendChild(empty);
    previous.disabled = true;
    next.disabled = true;
    return;
  }

  function draftLabel(data) {
    var base = "Draft";
    if (data.rejectedCount > 0) {
      base += " — " + data.rejectedCount + " claim" + (data.rejectedCount === 1 ? "" : "s") + " kept out";
    }
    return base + " · review before presenting";
  }

  function element(tag, className, text) {
    var node = document.createElement(tag);
    if (className) {
      node.className = className;
    }
    if (text !== undefined && text !== null) {
      node.textContent = text;
    }
    return node;
  }

  function buildClaim(claim) {
    var item = element("li", "claim");
    item.appendChild(element("p", "claim__text", claim.text));

    var sources = element("div", "sources");
    var sourcesId = "sources-" + claim.id;
    sources.id = sourcesId;
    sources.hidden = true;

    var button = element("button", "dd-code claim__prov", claim.evidenceIds.join(" "));
    button.type = "button";
    button.setAttribute("aria-expanded", "false");
    button.setAttribute("aria-controls", sourcesId);
    button.addEventListener("click", function () {
      var open = button.getAttribute("aria-expanded") === "true";
      button.setAttribute("aria-expanded", open ? "false" : "true");
      sources.hidden = open;
    });

    sources.appendChild(element("p", "dd-code sources__head", "Sources"));

    var list = element("ul", "sources__list");
    claim.evidenceIds.forEach(function (id) {
      var fact = factsById[id];
      var row = element("li", "sources__row");
      row.appendChild(element("span", "dd-code sources__id", id));
      row.appendChild(element("span", "sources__label", fact ? fact.label : id));
      row.appendChild(element("span", "dd-num sources__value", fact ? fact.value : ""));
      list.appendChild(row);
    });
    sources.appendChild(list);

    item.appendChild(button);
    item.appendChild(sources);
    return item;
  }

  function buildSlide(slide) {
    var section = element("section", "slide");
    section.id = "slide-" + slide.number;
    section.setAttribute("aria-label", "Slide " + slide.number + " of " + slides.length);
    if (slide.number !== 1) {
      section.hidden = true;
    }

    var inner = element("div", "slide__inner");
    inner.appendChild(element("h2", "slide__title", slide.title));

    var claims = element("ul", "claims");
    slide.claims.forEach(function (claim) {
      claims.appendChild(buildClaim(claim));
    });
    inner.appendChild(claims);

    if (slide.notes) {
      inner.appendChild(element("p", "slide__notes", slide.notes));
    }

    section.appendChild(inner);
    return section;
  }

  function show(index) {
    if (index < 0 || index >= slides.length) {
      return;
    }
    current = index;
    Array.prototype.forEach.call(slidesRoot.children, function (section, i) {
      section.hidden = i !== current;
    });
    Array.prototype.forEach.call(indexRoot.children, function (item, i) {
      item.firstChild.setAttribute("aria-current", i === current ? "true" : "false");
    });
    counter.textContent = "Slide " + (current + 1) + " of " + slides.length;
    previous.disabled = current === 0;
    next.disabled = current === slides.length - 1;
    document.getElementById("slide-" + slides[current].number).focus();
  }

  slides.forEach(function (slide) {
    slidesRoot.appendChild(buildSlide(slide));

    var item = document.createElement("li");
    var button = element("button", "dd-num", String(slide.number));
    button.type = "button";
    button.setAttribute("aria-label", "Slide " + slide.number);
    button.addEventListener("click", function () {
      show(slide.number - 1);
    });
    item.appendChild(button);
    indexRoot.appendChild(item);
  });

  previous.addEventListener("click", function () {
    show(current - 1);
  });
  next.addEventListener("click", function () {
    show(current + 1);
  });

  document.addEventListener("keydown", function (event) {
    if (event.altKey || event.metaKey || event.ctrlKey) {
      return;
    }
    switch (event.key) {
      case "ArrowRight":
      case "PageDown":
        show(current + 1);
        break;
      case "ArrowLeft":
      case "PageUp":
        show(current - 1);
        break;
      case "Home":
        show(0);
        break;
      case "End":
        show(slides.length - 1);
        break;
      default:
        return;
    }
    event.preventDefault();
  });

  show(0);
})();
