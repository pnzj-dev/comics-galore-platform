/**
 * Pure HTML + TypeScript gallery with inline main image, dots, thumbnail
 * grid, and a fullscreen lightbox on click. No framework dependencies —
 * styling is done entirely with default Tailwind utility classes, so you
 * just need the Tailwind CDN script (or your own Tailwind build) on the page.
 *
 * Usage:
 *   import { Gallery } from "./gallery";
 *
 *   new Gallery(document.getElementById("gallery")!, [
 *     { src: "photo1.jpg", alt: "...", caption: "..." },
 *     { src: "photo2.jpg", alt: "...", caption: "..." },
 *   ]);
 */

export interface GalleryImage {
  src: string;
  alt?: string;
  caption?: string;
}

const DOT_ACTIVE = ["w-4", "bg-white"];
const DOT_INACTIVE = ["w-1.5", "bg-white/30", "hover:bg-white/50"];

const THUMB_ACTIVE = ["opacity-100", "border-white"];
const THUMB_INACTIVE = ["opacity-60", "border-transparent", "hover:opacity-90"];

export class Gallery {
  private images: GalleryImage[];
  private index: number;
  private root: HTMLElement;

  // Inline gallery elements
  private mainImg!: HTMLImageElement;
  private dotsEl!: HTMLDivElement;
  private thumbsEl!: HTMLDivElement;

  // Fullscreen overlay elements (created lazily)
  private overlay: HTMLDivElement | null = null;
  private overlayImg: HTMLImageElement | null = null;
  private overlayCounter: HTMLSpanElement | null = null;
  private overlayCaption: HTMLParagraphElement | null = null;

  private touchStartX: number | null = null;

  constructor(root: HTMLElement, images: GalleryImage[], initialIndex = 0) {
    if (images.length === 0) throw new Error("Gallery requires at least one image");
    this.root = root;
    this.images = images;
    this.index = initialIndex;
    this.renderInline();
  }

  private get current(): GalleryImage {
    return this.images[this.index];
  }

  private goTo(next: number): void {
    this.index = (next + this.images.length) % this.images.length;
    this.updateInline();
    if (this.overlay) this.updateOverlay();
  }

  private goNext = (): void => this.goTo(this.index + 1);
  private goPrev = (): void => this.goTo(this.index - 1);

  /* ---------------- Inline gallery: main image + dots + thumbnails ---------------- */

  private renderInline(): void {
    this.root.className = "w-full max-w-xl mx-auto";
    this.root.innerHTML = "";

    // Main image wrapper
    const mainWrap = document.createElement("div");
    mainWrap.className = "relative aspect-[4/3] rounded-xl overflow-hidden bg-neutral-900";

    this.mainImg = document.createElement("img");
    this.mainImg.className = "w-full h-full object-cover cursor-zoom-in select-none";
    this.mainImg.alt = this.current.alt ?? "";
    this.mainImg.src = this.current.src;
    this.mainImg.draggable = false;
    this.mainImg.addEventListener("click", () => this.openFullscreen());
    mainWrap.appendChild(this.mainImg);

    if (this.images.length > 1) {
      mainWrap.appendChild(this.iconButton("prev", "Previous image", this.goPrev, "left-2.5"));
      mainWrap.appendChild(this.iconButton("next", "Next image", this.goNext, "right-2.5"));
    }

    this.root.appendChild(mainWrap);

    // Dots
    this.dotsEl = document.createElement("div");
    this.dotsEl.className = "flex justify-center gap-1.5 mt-3";
    this.root.appendChild(this.dotsEl);

    // Thumbnail grid — a direct child of the width-capped root, so it never
    // exceeds the main image's width
    this.thumbsEl = document.createElement("div");
    this.thumbsEl.className = "grid grid-cols-4 sm:grid-cols-5 gap-2 mt-3";
    this.root.appendChild(this.thumbsEl);

    this.renderDots();
    this.renderThumbs();
  }

  private renderDots(): void {
    if (this.images.length <= 1) return;
    this.dotsEl.innerHTML = "";
    this.images.forEach((_, i) => {
      const dot = document.createElement("button");
      dot.className = ["h-1.5", "rounded-full", "transition-all", ...(i === this.index ? DOT_ACTIVE : DOT_INACTIVE)].join(" ");
      dot.setAttribute("aria-label", `Go to image ${i + 1}`);
      dot.addEventListener("click", () => this.goTo(i));
      this.dotsEl.appendChild(dot);
    });
  }

  private renderThumbs(): void {
    if (this.images.length <= 1) return;
    this.thumbsEl.innerHTML = "";
    this.images.forEach((img, i) => {
      const btn = document.createElement("button");
      btn.className = [
        "aspect-square",
        "rounded-md",
        "overflow-hidden",
        "border-2",
        "transition-opacity",
        ...(i === this.index ? THUMB_ACTIVE : THUMB_INACTIVE),
      ].join(" ");
      btn.setAttribute("aria-label", `Show image ${i + 1}`);
      const thumbImg = document.createElement("img");
      thumbImg.className = "w-full h-full object-cover";
      thumbImg.src = img.src;
      thumbImg.alt = "";
      btn.appendChild(thumbImg);
      btn.addEventListener("click", () => this.goTo(i));
      this.thumbsEl.appendChild(btn);
    });
  }

  private updateInline(): void {
    this.mainImg.src = this.current.src;
    this.mainImg.alt = this.current.alt ?? "";

    this.dotsEl.querySelectorAll<HTMLButtonElement>("button").forEach((dot, i) => {
      dot.classList.remove(...DOT_ACTIVE, ...DOT_INACTIVE);
      dot.classList.add(...(i === this.index ? DOT_ACTIVE : DOT_INACTIVE));
    });
    this.thumbsEl.querySelectorAll<HTMLButtonElement>("button").forEach((thumb, i) => {
      thumb.classList.remove(...THUMB_ACTIVE, ...THUMB_INACTIVE);
      thumb.classList.add(...(i === this.index ? THUMB_ACTIVE : THUMB_INACTIVE));
    });
  }

  private iconButton(kind: "prev" | "next", label: string, onClick: () => void, position: string): HTMLButtonElement {
    const btn = document.createElement("button");
    btn.className = [
      "absolute",
      "top-1/2",
      "-translate-y-1/2",
      position,
      "flex",
      "items-center",
      "justify-center",
      "w-9",
      "h-9",
      "rounded-full",
      "bg-black/45",
      "text-white",
      "hover:bg-black/70",
      "transition-colors",
    ].join(" ");
    btn.setAttribute("aria-label", label);
    btn.innerHTML = kind === "prev" ? "&#10094;" : "&#10095;";
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      onClick();
    });
    return btn;
  }

  /* ---------------- Fullscreen overlay ---------------- */

  private openFullscreen(): void {
    if (!this.overlay) this.buildOverlay();
    this.updateOverlay();
    document.body.style.overflow = "hidden";
    this.overlay!.classList.remove("hidden");
    this.overlay!.classList.add("flex");
    this.overlay!.querySelector<HTMLButtonElement>("[data-close]")?.focus();
    window.addEventListener("keydown", this.onKeyDown);
  }

  private closeFullscreen = (): void => {
    if (!this.overlay) return;
    this.overlay.classList.add("hidden");
    this.overlay.classList.remove("flex");
    document.body.style.overflow = "";
    window.removeEventListener("keydown", this.onKeyDown);
  };

  private onKeyDown = (e: KeyboardEvent): void => {
    if (e.key === "Escape") this.closeFullscreen();
    if (e.key === "ArrowRight") this.goNext();
    if (e.key === "ArrowLeft") this.goPrev();
  };

  private buildOverlay(): void {
    const overlay = document.createElement("div");
    overlay.className = "fixed inset-0 z-50 hidden flex-col bg-black/95 backdrop-blur-sm";
    overlay.addEventListener("click", () => this.closeFullscreen());

    const topBar = document.createElement("div");
    topBar.className = "flex items-center justify-between px-4 py-3 text-white/80 shrink-0";
    topBar.addEventListener("click", (e) => e.stopPropagation());

    this.overlayCounter = document.createElement("span");
    this.overlayCounter.className = "text-sm tabular-nums select-none";

    const closeBtn = document.createElement("button");
    closeBtn.className = "p-2 rounded-full hover:bg-white/10 transition-colors";
    closeBtn.setAttribute("data-close", "");
    closeBtn.setAttribute("aria-label", "Close");
    closeBtn.innerHTML = "&#10005;";
    closeBtn.addEventListener("click", () => this.closeFullscreen());

    topBar.appendChild(this.overlayCounter);
    topBar.appendChild(closeBtn);

    const stage = document.createElement("div");
    stage.className = "relative flex-1 flex items-center justify-center overflow-hidden px-2";
    stage.addEventListener("click", (e) => e.stopPropagation());
    stage.addEventListener("touchstart", (e) => {
      this.touchStartX = e.touches[0].clientX;
    });
    stage.addEventListener("touchend", (e) => {
      if (this.touchStartX === null) return;
      const delta = e.changedTouches[0].clientX - this.touchStartX;
      if (Math.abs(delta) > 50) (delta > 0 ? this.goPrev : this.goNext)();
      this.touchStartX = null;
    });

    if (this.images.length > 1) {
      stage.appendChild(this.iconButton("prev", "Previous image", this.goPrev, "left-3"));
    }

    this.overlayImg = document.createElement("img");
    this.overlayImg.className = "max-h-full max-w-full object-contain select-none";
    this.overlayImg.draggable = false;
    stage.appendChild(this.overlayImg);

    if (this.images.length > 1) {
      stage.appendChild(this.iconButton("next", "Next image", this.goNext, "right-3"));
    }

    this.overlayCaption = document.createElement("p");
    this.overlayCaption.className = "shrink-0 text-center text-white/70 text-sm px-4 pb-4";
    this.overlayCaption.addEventListener("click", (e) => e.stopPropagation());

    overlay.appendChild(topBar);
    overlay.appendChild(stage);
    overlay.appendChild(this.overlayCaption);

    document.body.appendChild(overlay);
    this.overlay = overlay;
  }

  private updateOverlay(): void {
    if (!this.overlay || !this.overlayImg || !this.overlayCounter || !this.overlayCaption) return;
    this.overlayImg.src = this.current.src;
    this.overlayImg.alt = this.current.alt ?? "";
    this.overlayCounter.textContent = `${this.index + 1} / ${this.images.length}`;
    this.overlayCaption.textContent = this.current.caption ?? "";
    this.overlayCaption.style.display = this.current.caption ? "block" : "none";
  }
}
