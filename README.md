# Place Names API

Auto-suggests a list of place names based on a supplied prefix.

## Data generation

First run the **init-db** docker compose service: this will download placename data from Gov.UK's 
[ONS GeoPortal](https://geoportal.statistics.gov.uk/datasets/208d9884575647c29f0dd5a1184e711a/about) and unpack it into a SQLite3 database (into the `./data` folder).

A list of placenames was extracted from the sqlite db using the following query:

```sql
SELECT DISTINCT place23nm FROM place_names
```

This was then uploaded to ChatGPT along with the following prompt:

> Here's a list of over 62K placenames from around the UK. I want you to read the entire file,
> and produce a CSV of exactly the same length: `field 1` should be the placename from the text file;
> `field 2` should be a relevancy score ranging from 0.0 to 1.0.
>
> You should determine the relevancy based on your knowledge of the given place, how popular it is,
> or how likely someone would want to find out information about that place. So for example _"Edinburgh"_
> would score much higher than _"Edinbane"_. Likewise, it is likely that someone would be more interested
> in _"Nottingham"_ rather than _"Nottinghamshire"_. 
>
> Make sure you give a relevancy score for every line. Just do it all in one go - I dont want a sample
> of 50 or anything. Give me a link where i can download the CSV.

This results in the CSV file [./data/placenames_with_relevancy.csv](./data/placenames_with_relevancy.csv)
which is then loaded into a [trie structure](https://en.wikipedia.org/wiki/Trie) when the server starts.