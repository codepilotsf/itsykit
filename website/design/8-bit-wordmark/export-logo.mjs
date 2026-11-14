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
const w = Math.max(...cells.map(c => c[2])) - x0, h = Math.max(...cells.map(c => c[3])) - y0;
const pathData = (selection, dx = x0, dy = y0) => selection.map(([a, b, c, d]) => `M${a-dx} ${b-dy}h${c-a}v${d-b}h${a-c}Z`).join('');
const geometry = pathData(cells);
const theme = fs.readFileSync('site/css/theme.css', 'utf8');
const token = name => {
  const found = theme.match(new RegExp(`--${name}:\\s*(#[0-9a-fA-F]{6})`));
  if (!found) throw Error(`Missing theme token ${name}`);
  return found[1];
};
const navy = token('brand-2'), pale = token('hero-ink'), ice = token('accent');
const svg = (ink, label = 'ItsyWeb') => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${h}" role="img" aria-labelledby="title"><title id="title">${label}</title><path fill="${ink}" d="${geometry}"/></svg>\n`;
fs.writeFileSync('site/images/logo.svg', svg(navy));
fs.writeFileSync('site/images/logo-light.svg', svg(pale));

// The gap between y and W is explicit in the approved specimen.
const itsy = cells.filter(c => c[2] <= 1105);
const iw = Math.max(...itsy.map(c => c[2])) - x0;
const ih = Math.max(...itsy.map(c => c[3])) - y0;
const scale = 58 / iw, tx = 3, ty = (64 - ih * scale) / 2;
fs.writeFileSync('site/images/favicon.svg', `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-labelledby="title"><title id="title">Itsy</title><rect width="64" height="64" rx="8" fill="${pale}"/><path fill="${navy}" transform="translate(${tx} ${ty}) scale(${scale})" d="${pathData(itsy)}"/></svg>\n`);
const provenance = {
  origin: 'Owner-selected option 5, Small Assembly, using the lettering only; no separate mark.',
  source,
  approval: 'Use number 5 without the mark; retain site theme colours; favicon uses the Itsy lettering.',
  method: 'Font-independent rectilinear SVG contours reconstructed from long edges of the approved raster lettering. No font substitution or runtime image processing.',
  bounds: [x0, y0, x0 + w, y0 + h],
  itsyBounds: [x0, y0, x0 + iw, y0 + ih],
  size: { width: w, height: h },
  colors: { navy, pale, ice },
  approved: true
};
fs.writeFileSync('design/8-bit-wordmark/production.json', JSON.stringify(provenance, null, 2) + '\n');
console.log(JSON.stringify({ ...provenance.size, cells: cells.length, xStations: x.length, yStations: y.length, svgBytes: fs.statSync('site/images/logo.svg').size }));
