declare module '*.css' {
  const content: string;
  export default content;
}

// Obsidian loads plugins in a CommonJS environment.
declare const module: { exports: unknown };
