(function () {
  // For better minification.
  var doc = document
  var addEventListenerString = "addEventListener"
  var documentAddEventListener = doc[addEventListenerString]
  var currentScript = document.currentScript
  var loc = location
  var currentScriptDataset = currentScript.dataset
  var currentScriptUrl = new URL(currentScript.src)
  var scheme = loc.protocol;
  var global = globalThis
  var doFetch = global.fetch
  var history = global.history
  var visitorIdString = "visitorId"
  var targetString = "target"

  // Script options.
  //
  // URL of prisme instance.
  var prismeUrl = currentScriptDataset.prismeUrl || currentScriptUrl.origin
  var prismeApiEventsUrl = prismeUrl.concat("/api/v1/events")
  // Tracked website domain.
  var domain = currentScriptDataset.domain || loc.host;
  // Path of current page.
  var path = currentScriptDataset.path || loc.pathname;
  // Enable/disable manual tracking.
  var manual = (!!currentScriptDataset.manual && currentScriptDataset.manual !== "false") || false
  // Visitor ID.
  var visitorId = currentScriptDataset[visitorIdString];
  // Track outbound links.
  var trackOutboundLinks = currentScriptDataset.outboundLinks !== "false"
  // Track file downloads.
  var trackFileDownloads = currentScriptDataset.fileDownloads !== "false"
  // Status code.
  var statusCode = currentScriptDataset.status || "200"

  // State variables.
  var referrer = doc.referrer.replace(loc.host, domain);
  var pageviewCount = 0
  var trackingDisabled = localStorage.getItem("prismeAnalytics.tracking.enable") === "false"

  function defaultOptions(options) {
    if (!options) options = {}

    if (!options.domain) {
      // Ignore domain variable when manual tracking is enabled.
      if (manual) options.domain = loc.host
      else options.domain = domain
    }

    if (!options.status) options.status = statusCode

    if (!options.path) {
      // Ignore path variable when manual tracking is enabled.
      // Ignore path variable after first page view event.
      if (manual || pageviewCount > 1) options.path = loc.pathname
      else options.path = path
    }

    if (!options[visitorIdString]) options[visitorIdString] = visitorId

    options.url = scheme.concat("//", options.domain, options.path, loc.search)

    return options
  }

  function configureHeaders(options, headers) {
    headers["X-Prisme-Referrer"] = options.url
    if (options[visitorIdString]) {
      headers["X-Prisme-Visitor-Id"] = options[visitorIdString].toString()
    }
    return headers
  }

  function fetchDefaultOptions(options) {
    return Object.assign({}, {
      method: "POST",
      referrerPolicy: "no-referrer-when-downgrade",
      keepalive: true,
    }, options)
  }

  function shouldFollowLink(event, anchor) {
    // Another handler prevent default behavior.
    if (event.defaultPrevented) { return false }

    var targetsCurrentWindow = !anchor[targetString] || anchor[targetString].match(/^_(self|parent|top)$/i)
    var isRegularClick = !(event.ctrlKey || event.metaKey || event.shiftKey) && event.type === 'click'
    return targetsCurrentWindow && isRegularClick
  }


  function pageview(options) {
    if (trackingDisabled) return;
    pageviewCount++
    options = defaultOptions(options)

    doFetch(prismeApiEventsUrl.concat("/pageviews"), fetchDefaultOptions({
      headers: configureHeaders(options, {
        "X-Prisme-Document-Referrer": referrer,
        "X-Prisme-Status": options.status,
      }),
    }));

    referrer = options.url
  }

  function sendClickEvent(kind, url, options) {
    if (trackingDisabled) return Promise.resolve();
    options = defaultOptions(options)

    return doFetch(prismeApiEventsUrl.concat(kind), fetchDefaultOptions({
      headers: configureHeaders(options, {}),
      body: url
    }));
  }

  function handleLinkClickEvent(event) {
    // Ignore auxclick event with non middle button click or event target
    // isn't an element.
    if ((event.type === 'auxclick' && event.button !== 1) ||
      !(event[targetString] instanceof Element)) return

    var anchor = event[targetString].closest("a")
    if (!anchor) return
    var url = new URL(anchor.href || "", loc.origin)
    url.search = ""

    // Outbound links.
    if (trackOutboundLinks && url.host !== loc.host) {
      // Send event.
      sendClickEvent("/outbound-links", url)
    }

    // File downloads.
    if (trackFileDownloads && anchor.getAttribute("download") !== null) {
      return sendClickEvent("/file-downloads", url)
    }
  }

  if (!manual && (trackOutboundLinks || trackFileDownloads)) {
    documentAddEventListener('click', handleLinkClickEvent)
    documentAddEventListener('auxclick', handleLinkClickEvent)
  }

  var globalPrisme = {
    pageview: pageview,
    trigger(eventName, properties, options) {
      if (trackingDisabled) return;
      options = defaultOptions(options)

      doFetch(prismeUrl.concat("/api/v1/events/custom/", eventName), fetchDefaultOptions({
        headers: configureHeaders(options, {
          "Content-Type": "application/json",
        }),
        body: JSON.stringify(properties)
      }));
    },
  }
  global.prisme = globalPrisme

  // Filter ping attributes of anchor to prevent double outbound link
  // click / file download event
  doc.querySelectorAll('a[ping]').forEach(function(anchor) {
    anchor.ping = anchor.ping
      .split(" ")
      .filter((url) => !url.includes(prismeUrl))
      .join(" ")
  })

  // Manual tracking insn't enabled.
  if (!manual) {
    // Don't expose pageview function.
    delete globalPrisme.pageview

    // Trigger automatic pageview.
    pageview();

    // If website use a front end router, listen to push state and pop state
    // events to send pageview.
    if (history) {
      var pushState = history.pushState;
      history.pushState = function() {
        pushState.apply(history, arguments);
        pageview();
      }
      global[addEventListenerString]('popstate', pageview)
    }
  }
})();
