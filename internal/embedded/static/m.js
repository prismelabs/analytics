(function () {
  var ps = document.currentScript.dataset.prismeScheme || "https";
  var pd = document.currentScript.dataset.prismeDomain || "prismeanalytics.com";
  var s = location.protocol;
  var d = document.currentScript.dataset.domain || location.host;
  fetch(ps.concat("://", pd, "/api/v1/events/pageviews"), {
    method: "POST",
    headers: {
      "X-Prisme-Referrer": s.concat('//', d, location.pathname),
      "X-Prisme-Document-Referrer": document.referrer
    },
    referrerPolicy: "no-referrer-when-downgrade"
  });
})();
