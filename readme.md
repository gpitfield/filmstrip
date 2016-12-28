### Filmstrip

#### Directory structure
Image files within the source directory can be organized into directories, and Filmstrip will generate a "collection" for each directory, indexing it on the main navigation. Images can be sorted by using the prefix `_#_`, so for example _2_portrait.jpg will be given index value 2. Using the config value `root-collection`, you can designate which collection will be the root (main) collection.

You can specify a default copyright attribution using the `copyright` config value. It will be used for images that do not have EXIF copyright data.

`home-title` specifies the name to show on the left side of the home navigation.