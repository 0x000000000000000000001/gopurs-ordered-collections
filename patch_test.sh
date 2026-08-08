#!/bin/bash
sed -i.bak 's/log "filterWithKey gives submap"/log "filterWithKey gives submap"; quickCheck $ \\(TestMap s :: TestMap String Int) p -> Debug.trace s \\_ -> M.isSubmap (M.filterWithKey p s) s/' test/Test/Data/Map.purs
