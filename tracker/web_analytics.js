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
  var manual = (!!currentScriptDataset.manual && currentScriptDataset.manual !== "false") || false
  // Visitor ID.
  var visitorId = currentScriptDataset.visitorId;
  // Track outbound links.
  var outboundLinks = currentScriptDataset.outboundLinks !== "false"
  // Track file downloads.
  var fileDownloads = currentScriptDataset.fileDownloads !== "false"
  var extraDownloadsFileTypes = (currentScriptDataset.extraDownloadsFileTypes || "").split(",")

  // State variables.
  var referrer = document.referrer.replace(loc.host, domain);
  var pageviewCount = 1
  var global = globalThis
  var supportsKeepAlive = 'Request' in global && 'keepalive' in new Request('')
  var trackFileDownloadsTypes = [
    '7z',
    'avi',
    'csv',
    'dmg'
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

    if (!options.visitorId) options.visitorId = visitorId

    options.url = scheme.concat("//", options.domain, options.path, loc.search)

    return options
  }

  function configureHeaders(options, headers) {
    headers["Access-Control-Max-Age"] = 3600 // 1 hour
    headers["X-Prisme-Referrer"] = options.url

    if (options.visitorId) {
      headers["X-Prisme-Visitor-Id"] = options.visitorId.toString()
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

  function sendClick(options) {
    options = defaultOptions(options)

    return fetch(prismeUrl.concat("/api/v1/events/clicks"), {
      method: methodPost,
      headers: configureHeaders(options, {
        "Content-Type": "application/json",
      }),
      keepalive: true,
      referrerPolicy: referrerPolicy,
      body: JSON.stringify({ tag: options.tag, attr: options.attr })
    });
  }

  function sendClickEvent(event, url, options) {
      // Follow links only if keepalive isn't supported.
      var followed = supportsKeepAlive
      var followLink = () => {
        if (!followed) {
          followed = true
          global.location.assign(url)
        }
      }
      // Firefox stable doesn't support keepalive.
      if (!supportsKeepAlive) {
        event.preventDefault()
        setTimeout(followLink, 5000)
      }
      sendClick(options).finally(followLink)
  }

  function handleLinkClickEvent(event) {
    // Ignore auxclick event with non middle button click or event target
    // isn't an element.
    if ((event.type === 'auxclick' && event.button !== 1) ||
      !(event.target instanceof Element)) return

    var link = event.target.closest("a")
    if (!link) return
    var url = new URL(link.href || "", loc.origin)
    url.search = ""

    if (outboundLinks && url.host !== loc.host)
      sendClickEvent(event, url, { tag: "a", attr: url.href })

    console.log(link.getAttribute("download"), trackFileDownloadsTypes.includes(url.pathname.split('.').pop()))
    if (fileDownloads &&
      (link.getAttribute("download") !== null ||
        trackFileDownloadsTypes.includes(url.pathname.split('.').pop())))
      sendClickEvent(event, url, { tag: "a", attr: url.href })
  }

  if (outboundLinks) {
    document.addEventListener('click', handleLinkClickEvent)
    document.addEventListener('auxclick', handleLinkClickEvent)
  }

  global.prisme = {
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
    },
    click: sendClick,
  }

  // Manual tracking insn't enabled.
  if (!manual) {
    // Don't expose pageview function.
    delete global.prisme.pageview

    // Trigger automatic pageview.
    pageview();

    // If website use a front end router, listen to push state and pop state
    // events to send pageview.
    if (global.history) {
      var pushState = global.history.pushState;
      global.history.pushState = function() {
        pushState.apply(global.history, arguments);
        pageview();
      }
      global.addEventListener('popstate', pageview)
    }
  }
})();
