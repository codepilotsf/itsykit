// Export the owner-approved flat artwork as font-independent SVG geometry.
// Authoring utility only: shipped files are ready to serve, with no build step.
import fs from 'node:fs';
import zlib from 'node:zlib';
import path from 'node:path';
const root = process.cwd();
const source = '.impeccable/mocks/decision/round-2/loop.png';
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
function classify(x,y) {
  const i=(y*width+x)*channels, r=pixels[i], g=pixels[i+1], b=pixels[i+2];
  if (channels===4 && pixels[i+3]<128) return 0;
  if (g>145 && b-r>45 && g-r>20) return 2;
  return .2126*r+.7152*g+.0722*b<145?1:0;
}
function distance(p,a,b) {
  const dx=b[0]-a[0], dy=b[1]-a[1], t=dx||dy?Math.max(0,Math.min(1,((p[0]-a[0])*dx+(p[1]-a[1])*dy)/(dx*dx+dy*dy))):0;
  return Math.hypot(p[0]-a[0]-t*dx,p[1]-a[1]-t*dy);
}
function simplify(points,epsilon=.65) {
  if (points.length<3) return points;
  let max=0,index=0;
  for(let i=1;i<points.length-1;i++){const d=distance(points[i],points[0],points.at(-1));if(d>max){max=d;index=i;}}
  return max>epsilon?[...simplify(points.slice(0,index+1),epsilon).slice(0,-1),...simplify(points.slice(index),epsilon)]:[points[0],points.at(-1)];
}
function outline(region,color) {
  const [left,top,right,bottom]=region, w=right-left,h=bottom-top, base=w+1;
  const at=(x,y)=>x>=0&&x<w&&y>=0&&y<h&&classify(x+left,y+top)===color;
  const edges=new Map();
  const add=(x,y,xx,yy)=>{const k=y*base+x; if(!edges.has(k))edges.set(k,[]); edges.get(k).push(yy*base+xx);};
  for(let y=0;y<h;y++)for(let x=0;x<w;x++)if(at(x,y)){
    if(!at(x,y-1))add(x,y,x+1,y);
    if(!at(x+1,y))add(x+1,y,x+1,y+1);
    if(!at(x,y+1))add(x+1,y+1,x,y+1);
    if(!at(x-1,y))add(x,y+1,x,y);
  }
  const loops=[];
  while(edges.size){
    const first=edges.keys().next().value; let current=first, points=[];
    do{
      points.push([current%base,Math.floor(current/base)]);
      const list=edges.get(current);if(!list)throw Error('Open contour');
      const next=list.pop();if(!list.length)edges.delete(current);current=next;
    }while(current!==first);
    const area=Math.abs(points.reduce((sum,p,i)=>{const q=points[(i+1)%points.length];return sum+p[0]*q[1]-q[0]*p[1];},0)/2);
    if(area<12||points.length<8)continue;
    // Smooth single-pixel quantization before simplifying, keeping original proportions.
    points=points.map((p,i,a)=>[0,1].map(axis=>[-2,-1,0,1,2].reduce((s,k)=>s+a[(i+k+a.length)%a.length][axis],0)/5));
    let far=1;for(let i=2;i<points.length;i++)if(Math.hypot(points[i][0]-points[0][0],points[i][1]-points[0][1])>Math.hypot(points[far][0]-points[0][0],points[far][1]-points[0][1]))far=i;
    points=[...simplify(points.slice(0,far+1)).slice(0,-1),...simplify([...points.slice(far),points[0]]).slice(0,-1)];
    loops.push(points);
  }
  const n=v=>Math.round(v*100)/100;
  const svgPath=loops.map(points=>{
    const mid=(a,b)=>`${n((a[0]+b[0])/2)} ${n((a[1]+b[1])/2)}`;
    return `M${mid(points.at(-1),points[0])}`+points.map((p,i)=>`Q${n(p[0])} ${n(p[1])} ${mid(p,points[(i+1)%points.length])}`).join('')+'Z';
  }).join('');
  return svgPath;
}
function bounds(region){let x0=Infinity,y0=Infinity,x1=0,y1=0;for(let y=region[1];y<region[3];y++)for(let x=region[0];x<region[2];x++)if(classify(x,y)){x0=Math.min(x0,x);y0=Math.min(y0,y);x1=Math.max(x1,x+1);y1=Math.max(y1,y+1);}return[x0-3,y0-3,x1+3,y1+3];}
const full=bounds([50,190,1505,620]), mark=bounds([50,190,475,620]);
const navy='#142b43',blue='#a4dbff';
const geometry=region=>({navy:outline(region,1),blue:outline(region,2)});
function svg(region,geo,title,foreground=navy){
  const w=region[2]-region[0],h=region[3]-region[1];
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${h}" role="img" aria-labelledby="title"><title id="title">${title}</title><path fill="${foreground}" fill-rule="evenodd" d="${geo.navy}"/><path fill="${blue}" fill-rule="evenodd" d="${geo.blue}"/></svg>\n`;
}
const fullGeo=geometry(full),markGeo=geometry(mark);
fs.writeFileSync('site/images/logo.svg',svg(full,fullGeo,'Itsy Web'));
fs.writeFileSync('site/images/logo-light.svg',svg(full,fullGeo,'Itsy Web','#f5f9ff'));
fs.writeFileSync('site/images/mark.svg',svg(mark,markGeo,'Itsy Web woven-loop mark'));
const mw=mark[2]-mark[0],mh=mark[3]-mark[1],scale=52/Math.max(mw,mh),tx=(64-mw*scale)/2,ty=(64-mh*scale)/2;
fs.writeFileSync('site/images/favicon.svg',`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><title>Itsy Web</title><rect width="64" height="64" rx="14" fill="#f5f9ff"/><g transform="translate(${tx} ${ty}) scale(${scale})"><path fill="${navy}" fill-rule="evenodd" d="${markGeo.navy}"/><path fill="${blue}" fill-rule="evenodd" d="${markGeo.blue}"/></g></svg>\n`);
const provenance={origin:'Owner-approved woven-loop concept with the original rounded Itsy Web wordmark.',source,method:'Font-independent SVG contours from the approved flat artwork; pixel quantization smoothed and brand inks normalized.',approved:true};
fs.writeFileSync('design/itsy-web-logo/production.json',JSON.stringify({...provenance,fullBounds:full,markBounds:mark,wordmark:'Original approved lettering preserved as outlines',colors:{navy,blue},approved:true},null,2)+'\n');
console.log(JSON.stringify({width:full[2]-full[0],height:full[3]-full[1],markWidth:mw,markHeight:mh,svgBytes:fs.statSync('site/images/logo.svg').size}));
