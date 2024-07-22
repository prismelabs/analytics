(function () {
  // For better minification.
  var currentScript = document.currentScript
  var loc = location
  var currentScriptDataset = currentScript.dataset
  var currentScriptUrl = new URL(currentScript.src)
  var referrerPolicy ="no-referrer-when-downgrade"
  var methodPost = "POST"
  var scheme = loc.protocol;

  // Script options.
  //
  // URL of prisme instance.
  var prismeUrl = currentScriptDataset.prismeUrl || currentScriptUrl.origin
  // Tracked website domain.
  var domain = currentScriptDataset.domain || loc.host;
  // Path of current page.
  var path = currentScriptDataset.path || loc.pathname;
  // Enable/disable manual tracking.
  var manual = !!currentScriptDataset.manual || false
  // Visitor ID.
  var visitorId = currentScriptDataset.visitorId;

  // State variables.
  var referrer = document.referrer.replace(loc.host, domain);
  var pageviewCount = 1

  function defaultOptions(options) {
    if (!options) options = {}

    if (!options.domain) {
      // Ignore domain variable when manual tracking is enabled.
      if (manual) options.domain = loc.host
      else options.domain = domain
    }

    if (!options.path) {
      // Ignore path variable when manual tracking is enabled or this isn't
      // first pageview event and path variable value is outdated.
      if (manual || pageviewCount > 1) options.path = loc.pathname
      else options.path = path
    }

    if (!options.visitorId) options.visitorId = visitorId

    options.url = scheme.concat("//", options.domain, options.path, loc.search)

    return options
  }

  function configureHeaders(options, headers) {
    headers["Access-Control-Max-Age"] = 3600 // 1 hour
    headers["X-Prisme-Referrer"] = options.url

    if (options.visitorId) {
      headers["X-Prisme-Visitor-Id"] = options.visitorId.toString()
      if (options.anonymous == true) {
        headers["X-Prisme-Visitor-Anon"] = "1"
      }
    }

    return headers
  }

  function pageview(options) {
    options = defaultOptions(options)

    fetch(prismeUrl.concat("/api/v1/events/pageviews"), {
      method: methodPost,
      headers: configureHeaders(options, {
        "X-Prisme-Document-Referrer": referrer,
      }),
      keepalive: true,
      referrerPolicy: referrerPolicy
    });

    referrer = options.url
    pageviewCount++
  }

  window.prisme = {
    pageview: pageview,
    trigger(eventName, properties, options) {
      options = defaultOptions(options)

      fetch(prismeUrl.concat("/api/v1/events/custom/", eventName), {
        method: methodPost,
        headers: configureHeaders(options, {
          "Content-Type": "application/json",
        }),
        keepalive: true,
        referrerPolicy: referrerPolicy,
        body: JSON.stringify(properties)
      });
    }
  }

  // Manual tracking insn't enabled.
  if (!manual) {
    // Don't expose pageview function.
    delete window.prisme.pageview

    // Trigger automatic pageview.
    pageview();

    // If website use a front end router, listen to push state and pop state
    // events to send pageview.
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
