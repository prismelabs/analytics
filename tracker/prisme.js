(function () {
  var currentScript = document.currentScript
  var currentScriptDataset = currentScript.dataset
  var currentScriptUrl = new URL(currentScript.src)
  var prismeUrl = currentScriptDataset.prismeUrl || currentScriptUrl.origin
  var scheme = location.protocol;
  var domain = currentScriptDataset.domain || location.host;
  var referrer = document.referrer.replace(location.host, domain);
  var pageUrl = function() { return location.toString().replace(location.host, domain) }

  function pageview() {
    fetch(prismeUrl.concat("/api/v1/events/pageviews"), {
      method: "POST",
      headers: {
        "Access-Control-Max-Age": 3600,
        "X-Prisme-Referrer": pageUrl(),
        "X-Prisme-Document-Referrer": referrer,
      },
      keepalive: true,
      referrerPolicy: "no-referrer-when-downgrade",
    });

    referrer = pageUrl()
  }

  window.prisme = {
    trigger: function(eventName, properties) {
      fetch(prismeUrl.concat("/api/v1/events/custom/", eventName), {
        method: "POST",
        headers: {
          "Access-Control-Max-Age": 3600,
          "X-Prisme-Referrer": pageUrl(),
          "X-Prisme-Document-Referrer": referrer,
          "Content-Type": "application/json",
        },
        keepalive: true,
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
