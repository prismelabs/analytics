(function () {
  var currentScript = document.currentScript
  var currentScriptDataset = currentScript.dataset
  var currentScriptUrl = new URL(currentScript.src)
  var prismeUrl = currentScriptDataset.prismeUrl || currentScriptUrl.origin
  var scheme = location.protocol;
  var domain = currentScriptDataset.domain || location.host;
  var path = currentScriptDataset.path || location.pathname;
  var manual = !!currentScriptDataset.manual || false
  var referrer = document.referrer.replace(location.host, domain);

  // For better minification.
  var referrerPolicy ="no-referrer-when-downgrade"
  var method = "POST"
  var accessControlMaxAgeHeader = "Access-Control-Max-Age"
  var accessControlMaxAge = 3600
  var prismeReferrerHeader = "X-Prisme-Referrer"

  function defaultOptions(options) {
    if (!options) options = {}
    if (!("url" in options)) {
      if (!manual)
        options.url = location.origin.replace(location.host, domain) + path + location.search
      else options.url = location.toString().replace(location.host, domain)
    }

    return options
  }

  function pageview(options) {
    options = defaultOptions(options)

    fetch(prismeUrl.concat("/api/v1/events/pageviews"), {
      method: "POST",
      headers: {
        [accessControlMaxAgeHeader]: accessControlMaxAge,
        [prismeReferrerHeader]: options.url.toString(),
        "X-Prisme-Document-Referrer": referrer,
      },
      keepalive: true,
      referrerPolicy: referrerPolicy
    });

    referrer = options.url.toString()
  }

  window.prisme ={
    pageview,
    trigger(eventName, properties, options) {
      options = defaultOptions(options)

      fetch(prismeUrl.concat("/api/v1/events/custom/", eventName), {
        method: "POST",
        headers: {
          [accessControlMaxAgeHeader]: accessControlMaxAge,
          [prismeReferrerHeader]: options.url.toString(),
          "Content-Type": "application/json",
        },
        keepalive: true,
        referrerPolicy: referrerPolicy,
        body: JSON.stringify(properties)
      });
    }
  }

  if (!manual) {
    pageview();

    if (window.history) {
      var pushState = window.history.pushState;
      window.history.pushState = function() {
        pushState.apply(window.history, arguments);
        pageview();
      }
      window.addEventListener('popstate', pageview)
    }
  }
})();
