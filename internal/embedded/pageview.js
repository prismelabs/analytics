(function() {
	var r = new XMLHttpRequest();
	var u = '{SERVER_URL}/api/v1/events';
	var l = window.location
  var d = { k: "pageview", h: l.hostname, p: l.pathname, s: l.search }
	r.open('POST', u, true);
	r.setRequestHeader('Content-Type', 'application/json');
	r.send(JSON.stringify(d));
})()

