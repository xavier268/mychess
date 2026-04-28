Choses à améliorer 

Evaluation de la valeur intrinseque de la position :
* prendre en compte le nombre total de occupancies pour distinguer ouverture, milieu et fin de partie.
   * en ouverture, prime au centre
   * en milieu de partie, prime au nombre de cases accessibles des tours, cavaliers, fous et reine, prime aux pions passés
   * en fin de partie : prime/pénalité au nombre de cases accessibles au roi
* Utuliser des tables PST pour le score staique
  * A claculer de façon dynalique
  * Suppose de stocker le score avec la position,
  * Besoin d'enrif=chir le move avec la piece qui bouge pour recalculer facilement le score incrémental (DoMove/UndoMove)
* 

Interface browser :
 * ~~ajouter un mode "problème" (case à cocher).~~ DONE v0.5.0

