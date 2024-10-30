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
  var extraDownloadsFileTypes = (currentScriptDataset.extraDownloadsFileTypes || "").split(",")

  // State variables.
  var referrer = doc.referrer.replace(loc.host, domain);
  var pageviewCount = 1
  var supportsKeepAlive = 'Request' in global && 'keepalive' in new Request('')
  var trackFileDownloadsTypes = [
    '7z',
    'avi',
    'csv',
    'dmg',
    'docx',
    'exe',
    'gz',
    'key',
    'midi',
    'mov',
    'mp3',
    'mp4',
    'mpeg',
    'pdf',
    'pkg',
    'pps',
    'ppt',
    'pptx',
    'rar',
    'rtf',
    'txt',
    'wav',
    'wma',
    'wmv',
    'xlsx',
    'zip',
  ].concat(extraDownloadsFileTypes)

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

    if (!options[visitorIdString]) options[visitorIdString] = visitorId

    options.url = scheme.concat("//", options.domain, options.path, loc.search)

    return options
  }

  function configureHeaders(options, headers) {
    headers["Access-Control-Max-Age"] = 3600 // 1 hour
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
    options = defaultOptions(options)

    doFetch(prismeApiEventsUrl.concat("/pageviews"), fetchDefaultOptions({
      headers: configureHeaders(options, {
        "X-Prisme-Document-Referrer": referrer,
      }),
    }));

    referrer = options.url
    pageviewCount++
  }

  function sendClickEvent(kind, url, options) {
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
      var shouldFollowLinkManually = !supportsKeepAlive && shouldFollowLink(event, url)
      var followed = false
      function followLink() {
        if (!followed && shouldFollowLinkManually) {
          followed = true
          global.location.assign(url)
        }
      }

      if (shouldFollowLinkManually) {
        event.preventDefault()
        setTimeout(followLink, 5000)
      }

      // Send event.
      sendClickEvent("/outbound-link", url).finally(followLink)
    }

    // File downloads.
    if (trackFileDownloads &&
      (trackFileDownloadsTypes.includes(url.pathname.split(".").pop()) ||
        anchor.getAttribute("download") !== null)) {
      return sendClickEvent("/file-download", url)
    }
  }

  if (!manual && (trackOutboundLinks || trackFileDownloads)) {
    documentAddEventListener('click', handleLinkClickEvent)
    documentAddEventListener('auxclick', handleLinkClickEvent)
  }

  var globalPrisme = {
    pageview: pageview,
    trigger(eventName, properties, options) {
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
