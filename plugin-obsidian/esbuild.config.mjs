import esbuild from 'esbuild';

async function build() {
  try {
    await esbuild.build({
      bundle: true,
      entryPoints: ['src/main.ts'],
      format: 'cjs',
      platform: 'node',
      external: ['obsidian'],
      outfile: 'main.js',
      minify: true,
      sourcemap: true,
      loader: { '.ts': 'ts', '.css': 'text' },
      define: {
        'process.env.NODE_ENV': '"production"',
      },
    });
    console.log('✓ Build successful: main.js');
    
    // Create dist directory and copy files
    import('fs').then((fs) => {
      fs.mkdirSync('dist', { recursive: true });
      fs.writeFileSync('dist/main.js', fs.readFileSync('main.js', 'utf8'));
      fs.writeFileSync('dist/manifest.json', fs.readFileSync('manifest.json', 'utf8'));
    });
  } catch (err) {
    console.error('✗ Build failed:', err);
    process.exit(1);
  }
}

build();
