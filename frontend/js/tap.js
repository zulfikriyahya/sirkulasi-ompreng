// Penangkap reader HID (keyboard emulator) berdasarkan pola ketikan cepat.
const Tap = {
    ambangMs: 50,
    handler: null,         // fungsi(uid) saat tap valid; di-override oleh tab Kartu
    buf: "", last: 0,
};

document.addEventListener("keydown", e => {
    const now = performance.now();
    if (e.key === "Enter") {
        if (Tap.buf.length >= 8 && Tap.buf.length <= 10) {
            e.preventDefault(); e.stopPropagation();
            const uid = Tap.buf; Tap.buf = "";
            if (Tap.handler) Tap.handler(uid);
            return;
        }
        Tap.buf = ""; return;
    }
    if (/^\d$/.test(e.key)) {
        if (now - Tap.last > Tap.ambangMs) Tap.buf = "";
        Tap.buf += e.key; Tap.last = now;
    } else if (e.key.length === 1) {
        Tap.buf = "";
    }
}, true);

// Simulasi HID: kirim event keydown cepat ke dokumen
function simulasiHID(uid) {
    for (const ch of uid) document.dispatchEvent(new KeyboardEvent("keydown", { key: ch, bubbles: true, cancelable: true }));
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }));
}