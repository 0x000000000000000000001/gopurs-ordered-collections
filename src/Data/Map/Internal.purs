module Data.Map.Internal
  ( Map
  , empty
  , isEmpty
  , singleton
  , insert
  , insertWith
  , lookup
  , member
  , delete
  , keys
  , values
  , union
  , unionWith
  , intersection
  , intersectionWith
  , difference
  , size
  , pop
  , alter
  , update
  , checkValid
  , findMin
  , findMax
  , submap
  , filterKeys
  , fromFoldable
  ) where

import Prelude
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Data.List (List)
import Data.List as List
import Data.Foldable (class Foldable, foldl)
import Data.FoldableWithIndex (class FoldableWithIndex)

foreign import data Map :: Type -> Type -> Type

foreign import empty :: forall k v. Map k v

foreign import isEmpty :: forall k v. Map k v -> Boolean

foreign import singleton :: forall k v. k -> v -> Map k v

fromOrdering :: Ordering -> Int
fromOrdering = case _ of
  LT -> -1
  EQ -> 0
  GT -> 1

foreign import insertImpl :: forall k v. (k -> k -> Ordering) -> (Ordering -> Int) -> k -> v -> Map k v -> Map k v

insert :: forall k v. Ord k => k -> v -> Map k v -> Map k v
insert = insertImpl compare fromOrdering

foreign import insertWithImpl :: forall k v. (k -> k -> Ordering) -> (Ordering -> Int) -> (v -> v -> v) -> k -> v -> Map k v -> Map k v

insertWith :: forall k v. Ord k => (v -> v -> v) -> k -> v -> Map k v -> Map k v
insertWith = insertWithImpl compare fromOrdering

foreign import lookupImpl :: forall k v. (v -> Maybe v) -> Maybe v -> (k -> k -> Ordering) -> (Ordering -> Int) -> k -> Map k v -> Maybe v

lookup :: forall k v. Ord k => k -> Map k v -> Maybe v
lookup = lookupImpl Just Nothing compare fromOrdering

member :: forall k v. Ord k => k -> Map k v -> Boolean
member k m = case lookup k m of
  Nothing -> false
  Just _  -> true

foreign import deleteImpl :: forall k v. (k -> k -> Ordering) -> (Ordering -> Int) -> k -> Map k v -> Map k v

delete :: forall k v. Ord k => k -> Map k v -> Map k v
delete = deleteImpl compare fromOrdering

foreign import keysImpl :: forall k v. Map k v -> Array k
keys :: forall k v. Map k v -> List k
keys m = List.fromFoldable (keysImpl m)

foreign import valuesImpl :: forall k v. Map k v -> Array v
values :: forall k v. Map k v -> List v
values m = List.fromFoldable (valuesImpl m)

foreign import unionWithImpl :: forall k v. (k -> k -> Ordering) -> (Ordering -> Int) -> (v -> v -> v) -> Map k v -> Map k v -> Map k v
unionWith :: forall k v. Ord k => (v -> v -> v) -> Map k v -> Map k v -> Map k v
unionWith = unionWithImpl compare fromOrdering

union :: forall k v. Ord k => Map k v -> Map k v -> Map k v
union = unionWith const

foreign import intersectionWithImpl :: forall k a b c. (k -> k -> Ordering) -> (Ordering -> Int) -> (a -> b -> c) -> Map k a -> Map k b -> Map k c
intersectionWith :: forall k a b c. Ord k => (a -> b -> c) -> Map k a -> Map k b -> Map k c
intersectionWith = intersectionWithImpl compare fromOrdering

intersection :: forall k a b. Ord k => Map k a -> Map k b -> Map k a
intersection = intersectionWith const

foreign import differenceImpl :: forall k v w. (k -> k -> Ordering) -> (Ordering -> Int) -> Map k v -> Map k w -> Map k v
difference :: forall k v w. Ord k => Map k v -> Map k w -> Map k v
difference = differenceImpl compare fromOrdering

foreign import sizeImpl :: forall k v. Map k v -> Int

size :: forall k v. Map k v -> Int
size = sizeImpl
pop :: forall k v. Ord k => k -> Map k v -> Maybe (Tuple v (Map k v))
pop k m = case lookup k m of
  Nothing -> Nothing
  Just v -> Just (Tuple v (delete k m))

alter :: forall k v. Ord k => (Maybe v -> Maybe v) -> k -> Map k v -> Map k v
alter f k m = case f (lookup k m) of
  Nothing -> delete k m
  Just v -> insert k v m

update :: forall k v. Ord k => (v -> Maybe v) -> k -> Map k v -> Map k v
update f k m = alter (case _ of
  Nothing -> Nothing
  Just v -> f v) k m

foreign import checkValid :: forall k v. Map k v -> Boolean
foreign import findMinImpl :: forall k v. (forall a. a -> Maybe a) -> (forall a. Maybe a) -> Map k v -> Maybe { key :: k, value :: v }
findMin :: forall k v. Map k v -> Maybe { key :: k, value :: v }
findMin = findMinImpl Just Nothing

foreign import findMaxImpl :: forall k v. (forall a. a -> Maybe a) -> (forall a. Maybe a) -> Map k v -> Maybe { key :: k, value :: v }
findMax :: forall k v. Map k v -> Maybe { key :: k, value :: v }
findMax = findMaxImpl Just Nothing

foreign import submapImpl :: forall k v. (k -> k -> Ordering) -> (Ordering -> Int) -> Maybe k -> Maybe k -> Map k v -> Map k v
submap :: forall k v. Ord k => Maybe k -> Maybe k -> Map k v -> Map k v
submap = submapImpl compare fromOrdering

foreign import filterKeysImpl :: forall k v. (k -> Boolean) -> Map k v -> Map k v
filterKeys :: forall k v. (k -> Boolean) -> Map k v -> Map k v
filterKeys = filterKeysImpl

instance eqMap :: (Eq k, Eq v) => Eq (Map k v) where
  eq m1 m2 = size m1 == size m2 && keys m1 == keys m2 && values m1 == values m2

foreign import mapImpl :: forall k a b. (a -> b) -> Map k a -> Map k b

instance functorMap :: Functor (Map k) where
  map = mapImpl

foreign import foldlImpl :: forall k v z. (z -> k -> v -> z) -> z -> Map k v -> z
foreign import foldrImpl :: forall k v z. (z -> k -> v -> z) -> z -> Map k v -> z

instance foldableMap :: Foldable (Map k) where
  foldl f z m = foldlImpl (\acc _ v -> f acc v) z m
  foldr f z m = foldrImpl (\acc _ v -> f v acc) z m
  foldMap f m = foldl (\acc v -> acc <> f v) mempty m

instance foldableWithIndexMap :: FoldableWithIndex k (Map k) where
  foldlWithIndex f z m = foldlImpl (\acc k v -> f k acc v) z m
  foldrWithIndex f z m = foldrImpl (\acc k v -> f k v acc) z m
  foldMapWithIndex f m = foldlImpl (\acc k v -> acc <> f k v) mempty m

fromFoldable :: forall f k v. Ord k => Foldable f => f (Tuple k v) -> Map k v
fromFoldable = foldl (\m (Tuple k v) -> insert k v m) empty
