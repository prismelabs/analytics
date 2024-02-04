(function () {
	var ps = document.currentScript?.dataset?.prismeScheme || "https";
	var pd = document.currentScript?.dataset?.prismeDomain || "prismeanalytics.com";
	var s = location.protocol;
	var d = document.currentScript?.dataset?.domain || location.host;
	var referrer = document.referrer

  function pageview() {
    fetch(ps.concat("://", pd, "/api/v1/events/pageviews"), {
      method: "POST",
      headers: {
        "X-Prisme-Referrer": s.concat('//', d, location.pathname),
        "X-Prisme-Document-Referrer": referrer
      },
      referrerPolicy: "no-referrer-when-downgrade"
    });
  }

  pageview();

  if (window.history) {
    var pushState = window.history.pushState;
    window.history.pushState = function() {
      if (location.pathname === referrer) {
        return;
      }
      referrer = new URL(location.pathname, location.protocol.concat("//", d))
      pushState.apply(this, arguments);
      pageview();
    }
    window.addEventListener('popstate', pageview)
  }
})();
