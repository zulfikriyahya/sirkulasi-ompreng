// Pratinjau logo: dimuat saat aplikasi dibuka dan disegarkan setelah pilih/hapus logo.
const LogoPreview = {
  async muat() {
    const img = document.getElementById("logo-img");
    const kosong = document.getElementById("logo-kosong");
    if (!img || !kosong) return;
    try {
      const uri = await API.GetLogoPreview();
      img.classList.toggle("hidden", !uri);
      kosong.classList.toggle("hidden", !!uri);
      if (uri) img.src = uri;
    } catch (e) {}
  },
};

document.addEventListener("DOMContentLoaded", () => {
  LogoPreview.muat();
  const svc = window.go && window.go.service && window.go.service.Service;
  if (!svc) return;
  ["PilihLogo", "HapusLogo"].forEach((n) => {
    const asli = svc[n];
    if (typeof asli !== "function") return;
    svc[n] = async (...a) => {
      const r = await asli.apply(svc, a);
      LogoPreview.muat();
      return r;
    };
  });
});
