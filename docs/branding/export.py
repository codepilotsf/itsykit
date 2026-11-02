"""One-time, dependency-free vector export of the approved FleaCMS artwork.

PNG alpha outlines are traced without changing the mascot. Nunito outlines are
read through the machine's existing FreeType library. No runtime dependency.
"""
import ctypes as C
import json, math, struct, zlib
from pathlib import Path

OUT = Path(__file__).resolve().parent
SOURCE = OUT / 'approved-mascot.png'
FONT = OUT / 'nunito-800.woff2'
OUT.mkdir(parents=True, exist_ok=True)

def alpha_png(path):
    data=path.read_bytes(); pos=8; compressed=b''
    while pos<len(data):
        n=struct.unpack_from('>I',data,pos)[0]; kind=data[pos+4:pos+8]; chunk=data[pos+8:pos+8+n]; pos+=12+n
        if kind==b'IHDR': w,h,depth,color,_,_,interlace=struct.unpack('>IIBBBBB',chunk)
        if kind==b'IDAT': compressed+=chunk
    assert depth==8 and color==6 and interlace==0, (depth,color,interlace)
    raw=zlib.decompress(compressed); stride=w*4; rows=[]; prev=bytearray(stride); pos=0
    def paeth(a,b,c):
        p=a+b-c; da,db,dc=abs(p-a),abs(p-b),abs(p-c)
        return a if da<=db and da<=dc else b if db<=dc else c
    for y in range(h):
        filt=raw[pos]; pos+=1; row=bytearray(raw[pos:pos+stride]); pos+=stride
        for x in range(stride):
            a=row[x-4] if x>=4 else 0; b=prev[x]; c=prev[x-4] if x>=4 else 0
            predict=(0,a,b,(a+b)//2,paeth(a,b,c))[filt]
            row[x]=(row[x]+predict)&255
        rows.append(row); prev=row
    return w,h,[{x for x in range(w) if row[x*4+3]>=128} for row in rows]

def simplify(points, tolerance=.65):
    if len(points)<3: return points
    a,b=points[0],points[-1]; dx,dy=b[0]-a[0],b[1]-a[1]; den=math.hypot(dx,dy)
    ds=[abs(dy*p[0]-dx*p[1]+b[0]*a[1]-b[1]*a[0])/den if den else math.dist(a,p) for p in points]
    k=max(range(len(ds)),key=ds.__getitem__)
    if ds[k]>tolerance: return simplify(points[:k+1],tolerance)[:-1]+simplify(points[k:],tolerance)
    return [a,b]

w,h,rows=alpha_png(SOURCE); edges={}
def add(a,b): edges.setdefault(a,[]).append(b)
for y,row in enumerate(rows):
    for x in row:
        if y==0 or x not in rows[y-1]: add((x,y),(x+1,y))
        if x+1 not in row: add((x+1,y),(x+1,y+1))
        if y==h-1 or x not in rows[y+1]: add((x+1,y+1),(x,y+1))
        if x-1 not in row: add((x,y+1),(x,y))
loops=[]
while edges:
    start=min(edges); points=[start]; current=start
    while True:
        targets=edges[current]; nxt=targets.pop()
        if not targets: del edges[current]
        points.append(nxt); current=nxt
        if current==start: break
    area=abs(sum(a[0]*b[1]-a[1]*b[0] for a,b in zip(points,points[1:])))/2
    if area<12: continue
    mid=len(points)//2
    points=simplify(points[:mid+1])[:-1]+simplify(points[mid:])
    loops.append(points)
markpath=' '.join('M'+' L'.join(f'{x} {y}' for x,y in pts[:-1])+' Z' for pts in loops)
bounds=[min(x for pts in loops for x,y in pts), min(y for pts in loops for x,y in pts),max(x for pts in loops for x,y in pts),max(y for pts in loops for x,y in pts)]

class Vec(C.Structure): _fields_=[('x',C.c_long),('y',C.c_long)]
class Generic(C.Structure): _fields_=[('data',C.c_void_p),('finalizer',C.c_void_p)]
class BBox(C.Structure): _fields_=[('xmin',C.c_long),('ymin',C.c_long),('xmax',C.c_long),('ymax',C.c_long)]
class Face(C.Structure):
    _fields_=[('num_faces',C.c_long),('face_index',C.c_long),('flags',C.c_long),('style_flags',C.c_long),('num_glyphs',C.c_long),('family',C.c_char_p),('style',C.c_char_p),('n_sizes',C.c_int),('sizes',C.c_void_p),('n_charmaps',C.c_int),('charmaps',C.c_void_p),('generic',Generic),('bbox',BBox),('units',C.c_ushort),('ascender',C.c_short),('descender',C.c_short),('height',C.c_short),('max_advance_width',C.c_short),('max_advance_height',C.c_short),('underline_position',C.c_short),('underline_thickness',C.c_short),('glyph',C.c_void_p)]
class Metrics(C.Structure): _fields_=[(n,C.c_long) for n in ['width','height','horiBearingX','horiBearingY','horiAdvance','vertBearingX','vertBearingY','vertAdvance']]
class Slot(C.Structure): _fields_=[('library',C.c_void_p),('face',C.c_void_p),('next',C.c_void_p),('index',C.c_uint),('generic',Generic),('metrics',Metrics)]
class Outline(C.Structure): _fields_=[('n_contours',C.c_short),('n_points',C.c_short),('points',C.POINTER(Vec)),('tags',C.c_void_p),('contours',C.c_void_p),('flags',C.c_int)]
class Glyph(C.Structure): _fields_=[('library',C.c_void_p),('clazz',C.c_void_p),('format',C.c_uint),('advance',Vec),('outline',Outline)]
MOVE=C.CFUNCTYPE(C.c_int,C.POINTER(Vec),C.c_void_p)
LINE=MOVE
CONIC=C.CFUNCTYPE(C.c_int,C.POINTER(Vec),C.POINTER(Vec),C.c_void_p)
CUBIC=C.CFUNCTYPE(C.c_int,C.POINTER(Vec),C.POINTER(Vec),C.POINTER(Vec),C.c_void_p)
class Funcs(C.Structure): _fields_=[('move',MOVE),('line',LINE),('conic',CONIC),('cubic',CUBIC),('shift',C.c_int),('delta',C.c_long)]
ft=C.CDLL('libfreetype.so.6'); lib=C.c_void_p(); assert ft.FT_Init_FreeType(C.byref(lib))==0
faceptr=C.c_void_p(); assert ft.FT_New_Face(lib,str(FONT).encode(),0,C.byref(faceptr))==0
face=C.cast(faceptr,C.POINTER(Face)).contents
if face.flags & 256:
    weight=(C.c_long*1)(800*65536); assert ft.FT_Set_Var_Design_Coordinates(faceptr,1,weight)==0
else:
    assert face.style in [b'ExtraBold',b'Extra Bold'], face.style
assert face.units>0 and face.family.startswith(b'Nunito'), (face.units,face.family)
paths=[]; pen=0; previous=0; size=72; scale=size/face.units
ft.FT_Get_Char_Index.restype=C.c_uint
ft.FT_Get_Glyph.argtypes=[C.c_void_p,C.POINTER(C.c_void_p)]
for char in 'FleaCMS':
    idx=ft.FT_Get_Char_Index(faceptr,ord(char)); assert idx
    kern=Vec(); assert ft.FT_Get_Kerning(faceptr,previous,idx,2,C.byref(kern))==0
    pen+=kern.x*scale
    assert ft.FT_Load_Char(faceptr,ord(char),1|2|8)==0
    advance=C.cast(face.glyph,C.POINTER(Slot)).contents.metrics.horiAdvance
    glyph=C.c_void_p(); assert ft.FT_Get_Glyph(face.glyph,C.byref(glyph))==0
    outline=C.cast(glyph,C.POINTER(Glyph)).contents.outline; commands=[]
    def xy(p): return f'{p.contents.x*scale+pen+114:.3f} {85-p.contents.y*scale:.3f}'
    @MOVE
    def move(p,u):
        if commands: commands.append('Z')
        commands.append('M'+xy(p)); return 0
    @LINE
    def line(p,u): commands.append('L'+xy(p)); return 0
    @CONIC
    def conic(c,p,u): commands.append('Q'+xy(c)+' '+xy(p)); return 0
    @CUBIC
    def cubic(c1,c2,p,u): commands.append('C'+xy(c1)+' '+xy(c2)+' '+xy(p)); return 0
    funcs=Funcs(move,line,conic,cubic,0,0)
    assert ft.FT_Outline_Decompose(C.byref(outline),C.byref(funcs),None)==0
    commands.append('Z'); paths.append(' '.join(commands)); ft.FT_Done_Glyph(glyph)
    pen+=advance*scale-size*.035; previous=idx
font_units=face.units
ft.FT_Done_Face(faceptr); ft.FT_Done_FreeType(lib)

factor=120/w; xmin,ymin,xmax,ymax=[v*factor for v in bounds]
def svg(body,viewbox,color='#19241c',title='FleaCMS'):
    return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{viewbox}" role="img" aria-labelledby="title"><title id="title">{title}</title><g fill="{color}">{body}</g></svg>\n'
mascot=f'<path fill-rule="evenodd" transform="scale({factor:.9f})" d="{markpath}"/>'
word=''.join(f'<path d="{p}"/>' for p in paths)
vb=f'{xmin-2:.3f} {ymin-2:.3f} {114+pen+5-xmin:.3f} {ymax-ymin+4:.3f}'
(OUT/'logo.svg').write_text(svg(mascot+word,vb))
(OUT/'logo-light.svg').write_text(svg(mascot+word,vb,'#f4f8f2'))
edge=max(bounds[2]-bounds[0],bounds[3]-bounds[1])+64
cx=(bounds[0]+bounds[2])/2; cy=(bounds[1]+bounds[3])/2
markvb=f'{cx-edge/2:.1f} {cy-edge/2:.1f} {edge:.1f} {edge:.1f}'
(OUT/'mark.svg').write_text(svg(f'<path fill-rule="evenodd" d="{markpath}"/>',markvb,title='Little Skip, the FleaCMS mascot'))
favicon=(OUT/'mark.svg').read_text().replace('<g fill=', '<style>@media(prefers-color-scheme:dark){g{fill:#f4f8f2}}</style><g fill=')
(OUT/'favicon.svg').write_text(favicon)
(OUT/'export.json').write_text(json.dumps({'approved_source':'flea-logo-exploration f5b19f8, accepted by owner','font':'Nunito 800; outlined through existing FreeType','mascot':'Alpha silhouette, threshold 128/255; subpixel outline simplification 0.65 source pixels','source_dimensions':[w,h],'source_ink_bounds':bounds,'contours':len(loops),'logo_viewbox':vb,'glyph_advance':pen,'font_units':font_units},indent=2)+'\n')
print({'source':[w,h],'bounds':bounds,'contours':len(loops),'font_units':font_units,'word_width':pen,'bytes':len((OUT/'logo.svg').read_bytes())})
