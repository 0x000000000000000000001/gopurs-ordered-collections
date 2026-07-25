module Data.Map.Internal where

import Prelude
import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Data.List (List, fromFoldable)

foreign import data Map :: Type -> Type -> Type

foreign import empty :: forall k v. Map k v

foreign import isEmpty :: forall k v. Map k v -> Boolean

singleton :: forall k v. k -> v -> Map k v
singleton k v = insert k v empty

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
keys m = fromFoldable (keysImpl m)

foreign import valuesImpl :: forall k v. Map k v -> Array v
values :: forall k v. Map k v -> List v
values m = fromFoldable (valuesImpl m)

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

foreign import size :: forall k v. Map k v -> Int

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
