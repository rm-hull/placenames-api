CREATE TABLE IF NOT EXISTS place_names (
    id INTEGER PRIMARY KEY,
    placeid INTEGER, 
    place23cd TEXT,
    placesort TEXT,
    place23nm TEXT,
    splitind TEXT,
    descnm TEXT,
    ctyhistnm TEXT,
    cty61nm TEXT,
    cty91nm TEXT,
    ctyltnm TEXT,
    ctry23nm TEXT,
    cty23cd TEXT,
    cty23nm TEXT,
    lad61nm TEXT,
    lad61desc TEXT,
    lad91nm TEXT,
    lad91desc TEXT,
    lad23cd TEXT,
    lad23nm TEXT,
    lad23desc TEXT,
    ced23cd TEXT,
    wd23cd TEXT,
    par23cd TEXT,
    hlth23cd TEXT,
    hlth23nm TEXT,
    regd23cd TEXT,
    regd23nm TEXT,
    rgn23cd TEXT,
    rgn23nm TEXT,
    npark23cd TEXT,
    npark23nm TEXT,
    bua22cd TEXT,
    pcon23cd TEXT,
    pcon23nm TEXT,
    eer23cd TEXT,
    eer23nm TEXT,
    pfa23cd TEXT,
    pfa23nm TEXT,
    gridgb1m TEXT,
    gridgb1e TEXT,
    gridgb1n TEXT,
    grid1km TEXT,
    lat REAL,
    long REAL
);

.mode csv
.headers on
.import /app/data.csv place_names

CREATE VIRTUAL TABLE IF NOT EXISTS place_names_rtree USING rtree(
    id,
    minLat, maxLat,
    minLong, maxLong
);

INSERT INTO place_names_rtree(id, minLat, maxLat, minLong, maxLong)
SELECT id, lat, lat, long, long FROM place_names;

CREATE INDEX IF NOT EXISTS idx_placesort ON place_names(placesort);
