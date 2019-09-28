# PRR Laboratoire 1

## Objectifs

- Écrire un premier programme de complexité moyenne en Go (Golang) et se familiariser avec un environnement de programmation Go
- Réaliser ses premières communications par UDP et par diffusion partielle (multicast)

## Énoncé du problème

Nous souhaitons implémenter un algorithme simple permettant de synchroniser approximativement les
horloges locales des tâches d'une application répartie. Comme nous le savons, chaque site d'un
environnement réparti possède sa propre horloge système, mais aussi, cette horloge a un décalage et
une dérive qui lui est propre. Le but de notre algorithme est de rattraper ce décalage sans pour autant
corriger l'horloge du système.
