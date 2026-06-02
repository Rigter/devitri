const IMAGE_EXT = /\.(jpg|jpeg|png|gif|svg|webp|bmp|ico)$/i;

export function isImageFileName(name: string): boolean {
  return IMAGE_EXT.test(name);
}

export function resolveNoteAssetPath(notePath: string, assetSrc: string): string {
  const src = assetSrc.trim();
  if (src.startsWith('./')) {
    const relative = src.slice(2);
    const dir = notePath.includes('/') ? notePath.slice(0, notePath.lastIndexOf('/')) : '';
    return dir ? `${dir}/${relative}` : relative;
  }
  if (!src.includes('/') && notePath.includes('/')) {
    const dir = notePath.slice(0, notePath.lastIndexOf('/'));
    return `${dir}/${src}`;
  }
  return src;
}
