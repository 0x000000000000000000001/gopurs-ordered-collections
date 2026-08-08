module Data.Map
  ( module Data.Map.Internal
  , keys
  , SemigroupMap(..)
  ) where

import Prelude (class Eq, class Monoid, class Ord, class Semigroup, append, void, (<<<))

import Data.Map.Internal (Map, alter, any, anyWithKey, catMaybes, checkValid, delete, difference, empty, filter, filterKeys, filterWithKey, findMax, findMin, fromFoldable, fromFoldableWith, fromFoldableWithIndex, insert, insertWith, intersection, intersectionWith, isEmpty, isSubmap, lookup, lookupLE, lookupLT, lookupGE, lookupGT, mapMaybe, mapMaybeWithKey, member, pop, singleton, size, submap, union, unionWith, update, values, toUnfoldable, toUnfoldableUnordered)
import Data.Newtype (class Newtype)
import Data.Set (Set, fromMap)

-- | The set of keys of the given map.
-- | See also `Data.Set.fromMap`.
keys :: forall k v. Map k v -> Set k
keys = fromMap <<< void

-- | `SemigroupMap k v` provides a `Semigroup` instance for `Map k v` whose
-- | definition depends on the `Semigroup` instance for the `v` type.
-- | You should only use this type when you need `Data.Map` to have
-- | a `Semigroup` instance.
-- |
-- | ```purescript
-- | let
-- |   s :: forall key value. key -> value -> SemigroupMap key value
-- |   s k v = SemigroupMap (singleton k v)
-- |
-- | (s 1     "foo") <> (s 1     "bar") == (s 1  "foobar")
-- | (s 1 (First 1)) <> (s 1 (First 2)) == (s 1 (First 1))
-- | (s 1  (Last 1)) <> (s 1  (Last 2)) == (s 1  (Last 2))
-- | ```
newtype SemigroupMap k v = SemigroupMap (Map k v)

derive newtype instance eqSemigroupMap :: (Eq k, Eq v) => Eq (SemigroupMap k v)
derive instance newtypeSemigroupMap :: Newtype (SemigroupMap k v) _

instance semigroupSemigroupMap :: (Ord k, Semigroup v) => Semigroup (SemigroupMap k v) where
  append (SemigroupMap l) (SemigroupMap r) = SemigroupMap (unionWith append l r)

instance monoidSemigroupMap :: (Ord k, Semigroup v) => Monoid (SemigroupMap k v) where
  mempty = SemigroupMap empty
