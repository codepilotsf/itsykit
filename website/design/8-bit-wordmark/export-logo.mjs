// Export the owner-approved flat artwork as font-independent SVG geometry.
// Authoring utility only: shipped files are ready to serve, with no build step.
import fs from 'node:fs';
import zlib from 'node:zlib';
import path from 'node:path';
const root = process.cwd();
const source = 'design/logo-exploration-20261114/small-assembly.png';
const png = fs.readFileSync(path.join(root, source));
let width, height, channels, compressed = [];
for (let offset = 8; offset < png.length;) {
  const size = png.readUInt32BE(offset), kind = png.toString('ascii', offset + 4, offset + 8);
  const data = png.subarray(offset + 8, offset + 8 + size);
  if (kind === 'IHDR') {
    width = data.readUInt32BE(0); height = data.readUInt32BE(4);
    if (data[8] !== 8 || ![2, 6].includes(data[9]) || data[12] !== 0) throw Error('Expected non-interlaced RGB/RGBA PNG');
    channels = data[9] === 6 ? 4 : 3;
  }
  if (kind === 'IDAT') compressed.push(data);
  offset += size + 12;
}
const inflated = zlib.inflateSync(Buffer.concat(compressed));
const stride = width * channels, pixels = Buffer.alloc(stride * height);
function paeth(a, b, c) {
  const p = a + b - c, aa = Math.abs(p-a), bb = Math.abs(p-b), cc = Math.abs(p-c);
  return aa <= bb && aa <= cc ? a : bb <= cc ? b : c;
}
for (let y=0; y<height; y++) {
  const filter=inflated[y*(stride+1)];
  for (let x=0; x<stride; x++) {
    const a=x>=channels?pixels[y*stride+x-channels]:0;
    const b=y?pixels[(y-1)*stride+x]:0;
    const c=y&&x>=channels?pixels[(y-1)*stride+x-channels]:0;
    const prediction=[0,a,b,Math.floor((a+b)/2),paeth(a,b,c)][filter];
    pixels[y*stride+x]=(inflated[y*(stride+1)+1+x]+prediction)&255;
  }
}

// Retain only the selected lettering. The separate iw emblem is left of x=500.
const region = [500, 200, 1720, 580];
const [left, top, right, bottom] = region;
const at = (x, y) => {
  if (x < left || x >= right || y < top || y >= bottom) return false;
  const i = (y * width + x) * channels;
  return pixels[i] < 90 && pixels[i + 1] < 110 && pixels[i + 2] < 110;
};

// Cluster long contour edges, discarding raster antialiasing and one-pixel noise.
// Reconstruct square steps rather than smoothing this pixel lettering to curves.
const xs = new Map(), ys = new Map();
for (let y = top; y <= bottom; y++) for (let x = left; x <= right; x++) {
  if (at(x, y) !== at(x - 1, y)) xs.set(x, (xs.get(x) || 0) + 1);
  if (at(x, y) !== at(x, y - 1)) ys.set(y, (ys.get(y) || 0) + 1);
}
function stations(counts) {
  const groups = [];
  for (const entry of [...counts].filter(([, count]) => count >= 12).sort((a, b) => a[0] - b[0])) {
    const group = groups.at(-1);
    if (group && entry[0] - group.at(-1)[0] <= 4) group.push(entry);
    else groups.push([entry]);
  }
  return groups.map(group => Math.round(group.reduce((sum, [v, n]) => sum + v * n, 0) / group.reduce((sum, [, n]) => sum + n, 0)));
}
const x = stations(xs), y = stations(ys);
const cells = [];
for (let j = 0; j < y.length - 1; j++) for (let i = 0; i < x.length - 1; i++) {
  if (at(Math.floor((x[i] + x[i + 1]) / 2), Math.floor((y[j] + y[j + 1]) / 2))) {
    cells.push([x[i], y[j], x[i + 1], y[j + 1]]);
  }
}
if (!cells.length) throw Error('No selected lettering found');
const x0 = Math.min(...cells.map(c => c[0])), y0 = Math.min(...cells.map(c => c[1]));
const h = Math.max(...cells.map(c => c[3])) - y0;
const pathData = (selection, dx = x0, dy = y0) => selection.map(([a, b, c, d]) => `M${a-dx} ${b-dy}h${c-a}v${d-b}h${a-c}Z`).join('');
// Preserve the approved Itsy letterforms; draw Kit in the same square-step grammar.
// These are font-independent contours, never a substituted bitmap font.
const itsy = cells.filter(c => c[2] <= 1105);
const iw = Math.max(...itsy.map(c => c[2])) - x0;
const kitX = iw + 18;
const k = 'M0 0H50V88H73V65H96V42H119V19H142V0H193V42H170V65H147V88H124V111H101V133H124V155H147V177H170V199H193V221H142V199H119V177H96V155H73V133H50V221H0Z';
const i = 'M0 0H50V47H0ZM0 68H50V221H0Z';
const tCells = cells.filter(c => c[0] >= 651 && c[2] <= 768);
const w = kitX + 193 + 18 + 50 + 18 + 117;
const geometry = pathData(itsy);
const kitGeometry = `<g transform="translate(${kitX} 0)"><path d="${k}"/><path transform="translate(211 0)" d="${i}"/><path transform="translate(279 0)" d="${pathData(tCells, 651, y0)}"/></g>`;
const theme = fs.readFileSync('site/css/theme.css', 'utf8');
const token = name => {
  const found = theme.match(new RegExp(`--${name}:\\s*(#[0-9a-fA-F]{6})`));
  if (!found) throw Error(`Missing theme token ${name}`);
  return found[1];
};
const navy = token('brand-2'), pale = token('hero-ink'), ice = token('accent');
const svg = ink => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${h}" role="img" aria-labelledby="title"><title id="title">ItsyKit</title><g fill="${ink}"><path d="${geometry}"/>${kitGeometry}</g></svg>\n`;
fs.writeFileSync('site/images/logo.svg', svg(navy));
fs.writeFileSync('site/images/logo-light.svg', svg(pale));

const ih = Math.max(...itsy.map(c => c[3])) - y0;
const scale = 58 / iw, tx = 3, ty = (64 - ih * scale) / 2;
fs.writeFileSync('site/images/favicon.svg', `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-labelledby="title"><title id="title">Itsy</title><rect width="64" height="64" rx="8" fill="${pale}"/><path fill="${navy}" transform="translate(${tx} ${ty}) scale(${scale})" d="${pathData(itsy)}"/></svg>\n`);
const provenance = {
  origin: 'Owner-selected option 5, Small Assembly, renamed ItsyKit with lettering only; no separate mark.',
  source,
  approval: 'Rename the brand ItsyKit, retain the 8-bit lettering and theme colours, and make the logo a little smaller. Favicon remains Itsy.',
  method: 'Approved Itsy contours reconstructed from the raster; Kit drawn as matching square-step SVG letterforms, reusing the approved t. No font substitution or runtime image processing.',
  bounds: [0, 0, w, h],
  itsyBounds: [x0, y0, x0 + iw, y0 + ih],
  size: { width: w, height: h },
  colors: { navy, pale, ice },
  approved: true
};
fs.writeFileSync('design/8-bit-wordmark/production.json', JSON.stringify(provenance, null, 2) + '\n');
console.log(JSON.stringify({ ...provenance.size, cells: cells.length, xStations: x.length, yStations: y.length, svgBytes: fs.statSync('site/images/logo.svg').size }));
