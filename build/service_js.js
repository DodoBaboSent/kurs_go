const staticCacheName = "s-app-v1"

const assetUrls = [
    '/',
    '/assets/main.wasm',
    '/assets/app.js',
    '/assets/app.css',
    '/static/offline.html',
    '/static/manifest.json',
    '/news',
    '/static/loopa.png'
]

self.addEventListener('install', event => {
    console.log("installed")
    event.waitUntil(
        caches.open(staticCacheName).then(cache => cache.addAll(assetUrls))
    )
})

self.addEventListener('activate', async event => {
    console.log("activated")
    const cacheNames = await caches.keys()
    await Promise.all(
        cacheNames
        .filter(name => name !== staticCacheName)
        .map(name => caches.delete(name))
    )
    return self.clients.claim();
})

self.addEventListener('fetch', async event => {
    const {request} = event
    
    const url = new URL(request.url)
    if (url.pathname === "/news") {
        event.respondWith(networkFirst(request))
    } else if (url.origin !== location.origin) {
        event.respondWith(networkFirst(request))
    } else {
        event.respondWith(cacheFirst(request))
    }
})


async function networkFirst(request){
    const cache = await caches.open(staticCacheName)
    try {
        const response = await fetch(request)
        await cache.put(request, response.clone())
        return response
    } catch (e){
        const cached = await cache.match(request)
        return cached ?? await cache.match("/static/offline.html")
    }

}

async function cacheFirst(request){
    const cache = await caches.open(staticCacheName)
    try{
        const cached = await cache.match(request)
        return cached ?? await fetch(request)
    } catch (e){
        return await cache.match("/static/offline.html")
    }
}