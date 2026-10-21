// Binding Wails: window.go.<package>.<Type>.<Method>
const API = new Proxy({}, {
    get: (_, name) => (...args) => {
        const svc = window.go && window.go.service && window.go.service.Service;
        if (!svc || !svc[name]) return Promise.reject(new Error("Binding belum siap: " + name));
        return svc[name](...args);
    }
});

let _audio;
function bunyi(tipe) {
    try {
        _audio = _audio || new (window.AudioContext || window.webkitAudioContext)();
        const o = _audio.createOscillator(), g = _audio.createGain();
        o.frequency.value = tipe === "ok" ? 880 : tipe === "err" ? 220 : 520;
        g.gain.value = 0.15;
        o.connect(g); g.connect(_audio.destination);
        o.start(); o.stop(_audio.currentTime + (tipe === "err" ? 0.4 : 0.15));
    } catch (e) { }
}

let _toastT;
function toast(teks, tipe = "info") {
    const el = document.getElementById("toast");
    el.textContent = teks; el.className = tipe;
    bunyi(tipe);
    clearTimeout(_toastT);
    _toastT = setTimeout(() => el.className = "hidden", 2500);
}