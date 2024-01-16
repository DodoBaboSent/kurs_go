const dynamicCacheName = "d-app-v1"

const assetUrls = [
    '/',
    '/assets/main.wasm',
    '/assets/app.js',
    '/assets/app.css',
    '/static/offline.html',
    '/static/manifest.json',
]

self.addEventListener('install', event => {
    console.log("installed")
    event.waitUntil(
        caches.open(dynamicCacheName).then(cache => cache.addAll(assetUrls))
    )
})

self.addEventListener('activate', async event => {
    console.log("activated")
    const cacheNames = await caches.keys()
    await Promise.all(
        cacheNames
        .filter(name => name !== dynamicCacheName)
        .map(name => caches.delete(name))
    )
    return self.clients.claim();
})

self.addEventListener('fetch', async event => {
    const {request} = event
    
    console.log("Fetch " + request.url)

    console.log("Going online")
    event.respondWith(networkFirst(request))

})


async function networkFirst(request){
    const cache = await caches.open(dynamicCacheName)
    try {
        const response = await fetch(request)
        await cache.put(request, response.clone())
        return response
    } catch (e){
        const cached = await cache.match(request)
        return cached ?? await cache.match("/static/offline.html")
    }

}