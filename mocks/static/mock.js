const PRISME_URL = process.env.PRISME_URL
const PORT = Number.parseInt(process.env.PORT) || 8000

let verificationId = ''

const server = Bun.serve({
  port: PORT,
  async fetch (request) {
    console.log(
      `${new Date().toISOString()} - ${request.method} ${request.url}`
    )
    if (request.method === 'POST') {
      verificationId = await request.json()
    }

    return new Response(indexHtml(verificationId), {
      headers: { 'Content-Type': 'text/html' }
    })
  }
})

function indexHtml (verificationId) {
  return `
<!doctype html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <link rel="icon" type="image/svg+xml" href="/vite.svg" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <script src="${PRISME_URL}/static/m.js" data-prisme-verification-id="${verificationId}" defer></script>
  <title>Static site</title>
  <script>
    document.addEventListener('DOMContentLoaded', () => {
      document.addEventListener('click', (ev) => {
        window.prisme.trigger('click', { x: ev.clientX, y: ev.clientY })
      })
    })
  </script>
</head>

<body>
  <h1>Index</h1>
  <a href="/">Home</a>
  <a href="/page1">Page 1</a>
  <a href="/page2">Page 2</a>
</body>

</html>`
}

console.log(`Listening on localhost: ${server.port}`)

