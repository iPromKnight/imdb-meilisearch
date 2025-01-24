# POC

Initial time boxed poc to demonstrate what I was talking about in discord.
This is a meilisearch powered title matcher for imdb dataset data, to match filenames 
or title / category and year to an imdb entry and score it.

There is an api allowing to query search endpoints for the data, or you can use it solely as a cli tool.

Running the docker compose file will setup a container running MS internally, it will seed the imdb data from the public none commercial dataset, and add the tool.

Go 1.23 is required. 

Created for educational purposes, and to demonstrate probably the fastest way for us to title match and score.



<a href='https://ko-fi.com/W7W616IBNG' target='_blank'><img height='36' style='border:0px;height:36px;' src='https://storage.ko-fi.com/cdn/kofi5.png?v=6' border='0' alt='Buy Me a Coffee at ko-fi.com' /></a>
