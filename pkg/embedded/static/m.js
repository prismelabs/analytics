(function () {
  var ssrc = new URL(document.currentScript.src)
  var ps = document.currentScript.dataset.prismeScheme || ssrc.protocol.slice(0, -1)
  var pd = document.currentScript.dataset.prismeDomain || ssrc.host
  var s = location.protocol;
  var d = document.currentScript.dataset.domain || location.host;
  var r = document.referrer.replace(location.host, d);

  function pageview() {
    var u = s.concat('//', d, location.pathname)

    fetch(ps.concat("://", pd, "/api/v1/events/pageviews"), {
      method: "POST",
      headers: {
        "X-Prisme-Referrer": u,
        "X-Prisme-Document-Referrer": r,
      },
      referrerPolicy: "no-referrer-when-downgrade"
    });

    r = u
  }

  pageview();

  if (window.history) {
    var pushState = window.history.pushState;
    window.history.pushState = function() {
      pushState.apply(window.history, arguments);
      pageview();
    }
    window.addEventListener('popstate', pageview)
  }
})();
