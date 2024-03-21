(function () {
  var currentScript = document.currentScript
  var currentScriptDataset = currentScript.dataset
  var currentScriptUrl = new URL(currentScript.src)
  var prismeApi = currentScriptDataset.prismeApi || currentScriptUrl.origin.concat("/api/v1")
  var scheme = location.protocol;
  var domain = currentScriptDataset.domain || location.host;
  var referrer = document.referrer.replace(location.host, domain);
  var pageUrl = function() { return location.toString().replace(location.host, domain) }

  function pageview() {
    fetch(prismeApi.concat("/events/pageviews"), {
      method: "POST",
      headers: {
        "Access-Control-Max-Age": 3600,
        "X-Prisme-Referrer": pageUrl(),
        "X-Prisme-Document-Referrer": referrer,
      },
      referrerPolicy: "no-referrer-when-downgrade"
    });

    referrer = pageUrl()
  }

  window.prisme = {
    trigger: function(eventName, properties) {
      fetch(prismeApi.concat("/events/customs/", eventName), {
        method: "POST",
        headers: {
          "Access-Control-Max-Age": 3600,
          "X-Prisme-Referrer": pageUrl(),
          "Content-Type": "application/json",
        },
        referrerPolicy: "no-referrer-when-downgrade",
        body: JSON.stringify(properties)
      });
    }
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
