#!/usr/bin/env python3
"""Build es_ru and en_ru picture quests JSON from original git source.

Loads /tmp/orig_pq.json, keeps Russian titles, fixes misaligned
completion_criteria for five quests, writes:
  - resources/picture_quests/es_ru_picture_quests_50.json (Spanish criteria)
  - resources/picture_quests/en_ru_picture_quests_50.json (English criteria)
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
INPUT_PATH = Path("/tmp/orig_pq.json")
OUTPUT_ES_PATH = REPO_ROOT / "resources/picture_quests/es_ru_picture_quests_50.json"
OUTPUT_EN_PATH = REPO_ROOT / "resources/picture_quests/en_ru_picture_quests_50.json"

CRITERIA_FIXES: dict[str, dict[str, str]] = {
  "mesa_taza_libro_manzana_llaves": {
    "actions": "Task is complete when the learner, in Spanish, mentions that there are no people in the scene. Minor grammar mistakes are acceptable if the meaning is clear.",
    "colors_numbers": "Task is complete when the learner, in Spanish, mentions at least two colors: white cup, blue book, red apple, silver keys. Minor grammar mistakes are acceptable if the meaning is clear.",
    "layout": "Task is complete when the learner, in Spanish, mentions where at least two objects are: cup on the left, book in the center, apple on the right, or keys in front; optionally the apple has a green leaf. Minor grammar mistakes are acceptable if the meaning is clear."
  },
  "habitacion_cama_lampara_ventana": {
    "actions": "Task is complete when the learner, in Spanish, mentions that there are no people in the room. Minor grammar mistakes are acceptable if the meaning is clear.",
    "colors_numbers": "Task is complete when the learner, in Spanish, mentions at least two colors: blue bed, yellow lamp, brown table, green rug, or white curtains. Minor grammar mistakes are acceptable if the meaning is clear.",
    "layout": "Task is complete when the learner, in Spanish, mentions where the bed, lamp, or window is located; optionally the green rug is on the floor. Minor grammar mistakes are acceptable if the meaning is clear."
  },
  "parada_autobus_un_autobus": {
    "actions": "Task is complete when the learner, in Spanish, mentions that the bus is stopped or that people are waiting at the stop. Minor grammar mistakes are acceptable if the meaning is clear.",
    "colors_numbers": "Task is complete when the learner, in Spanish, mentions that two people are waiting and at least one color such as blue bus, gray shelter, or blue sky. Minor grammar mistakes are acceptable if the meaning is clear.",
    "layout": "Task is complete when the learner, in Spanish, mentions where the bus or shelter is: bus on the right, people under the shelter on the left. Minor grammar mistakes are acceptable if the meaning is clear."
  },
  "playa_sombrilla_toalla_mar": {
    "actions": "Task is complete when the learner, in Spanish, mentions that it is sunny or a beach scene. Minor grammar mistakes are acceptable if the meaning is clear.",
    "colors_numbers": "Task is complete when the learner, in Spanish, mentions at least two colors: blue sea, yellow umbrella, red towel, green sandals. Minor grammar mistakes are acceptable if the meaning is clear."
  },
  "desayuno_pan_leche_banana": {
    "actions": "Task is complete when the learner, in Spanish, mentions that there are no people and that breakfast is on the table. Minor grammar mistakes are acceptable if the meaning is clear.",
    "colors_numbers": "Task is complete when the learner, in Spanish, mentions the quantities (two slices of bread, one glass of milk, or one banana) and at least two colors: brown bread, white milk, yellow banana, white plate. Minor grammar mistakes are acceptable if the meaning is clear."
  }
}

SPANISH_CRITERIA: dict[str, str] = {
  "mesa_taza_libro_manzana_llaves::main_objects": "La tarea se completa cuando el alumno, en español, nombra la taza, el libro, la manzana y las llaves, y menciona la mesa. Se aceptan errores gramaticales menores si el significado es claro.",
  "mesa_taza_libro_manzana_llaves::actions": "La tarea se completa cuando el alumno, en español, menciona que no hay personas en la escena. Se aceptan errores gramaticales menores si el significado es claro.",
  "mesa_taza_libro_manzana_llaves::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores: taza blanca, libro azul, manzana roja o llaves plateadas. Se aceptan errores gramaticales menores si el significado es claro.",
  "mesa_taza_libro_manzana_llaves::layout": "La tarea se completa cuando el alumno, en español, menciona dónde están al menos dos objetos: la taza a la izquierda, el libro en el centro, la manzana a la derecha o las llaves delante; opcionalmente, que la manzana tiene una hoja verde. Se aceptan errores gramaticales menores si el significado es claro.",
  "habitacion_cama_lampara_ventana::main_objects": "La tarea se completa cuando el alumno, en español, nombra la cama, la lámpara, la ventana y la habitación. Se aceptan errores gramaticales menores si el significado es claro.",
  "habitacion_cama_lampara_ventana::actions": "La tarea se completa cuando el alumno, en español, menciona que no hay personas en la habitación. Se aceptan errores gramaticales menores si el significado es claro.",
  "habitacion_cama_lampara_ventana::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores: cama azul, lámpara amarilla, mesita marrón, alfombra verde o cortinas blancas. Se aceptan errores gramaticales menores si el significado es claro.",
  "habitacion_cama_lampara_ventana::layout": "La tarea se completa cuando el alumno, en español, menciona dónde están la cama, la lámpara o la ventana; opcionalmente, que la alfombra verde está en el suelo. Se aceptan errores gramaticales menores si el significado es claro.",
  "gato_silla_pelota::main_objects": "La tarea se completa cuando el alumno, en español, menciona el gato, la silla y la pelota. Se aceptan errores gramaticales menores si el significado es claro.",
  "gato_silla_pelota::actions": "La tarea se completa cuando el alumno, en español, menciona que el gato está en la silla. Se aceptan errores gramaticales menores si el significado es claro.",
  "gato_silla_pelota::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos un color: gato naranja, pelota azul, silla de madera o suelo beige. Se aceptan errores gramaticales menores si el significado es claro.",
  "gato_silla_pelota::layout": "La tarea se completa cuando el alumno, en español, menciona que la pelota está a la derecha de la silla. Se aceptan errores gramaticales menores si el significado es claro.",
  "nina_globo_rojo::main_objects": "La tarea se completa cuando el alumno, en español, menciona la niña y el globo rojo. Se aceptan errores gramaticales menores si el significado es claro.",
  "nina_globo_rojo::actions": "La tarea se completa cuando el alumno, en español, menciona que la niña sostiene el globo. Se aceptan errores gramaticales menores si el significado es claro.",
  "nina_globo_rojo::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores: globo rojo, vestido amarillo, zapatos azules, hierba verde o cielo azul. Se aceptan errores gramaticales menores si el significado es claro.",
  "nina_globo_rojo::layout": "La tarea se completa cuando el alumno, en español, menciona el árbol detrás de la niña. Se aceptan errores gramaticales menores si el significado es claro.",
  "desayuno_pan_leche_banana::main_objects": "La tarea se completa cuando el alumno, en español, menciona el pan, la leche y el plátano. Se aceptan errores gramaticales menores si el significado es claro.",
  "desayuno_pan_leche_banana::actions": "La tarea se completa cuando el alumno, en español, menciona que no hay personas y que el desayuno está en la mesa. Se aceptan errores gramaticales menores si el significado es claro.",
  "desayuno_pan_leche_banana::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona las cantidades (dos rebanadas de pan, un vaso de leche o un plátano) y al menos dos colores: pan marrón, leche blanca, plátano amarillo o plato blanco. Se aceptan errores gramaticales menores si el significado es claro.",
  "desayuno_pan_leche_banana::layout": "La tarea se completa cuando el alumno, en español, menciona dónde está el vaso o el plátano. Se aceptan errores gramaticales menores si el significado es claro.",
  "parada_autobus_un_autobus::main_objects": "La tarea se completa cuando el alumno, en español, menciona el autobús, la parada y las personas. Se aceptan errores gramaticales menores si el significado es claro.",
  "parada_autobus_un_autobus::actions": "La tarea se completa cuando el alumno, en español, menciona que el autobús está parado o que las personas esperan en la parada. Se aceptan errores gramaticales menores si el significado es claro.",
  "parada_autobus_un_autobus::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona que dos personas esperan y al menos un color, como autobús azul, refugio gris o cielo azul. Se aceptan errores gramaticales menores si el significado es claro.",
  "parada_autobus_un_autobus::layout": "La tarea se completa cuando el alumno, en español, menciona dónde están el autobús o el refugio: el autobús a la derecha, las personas bajo el refugio a la izquierda. Se aceptan errores gramaticales menores si el significado es claro.",
  "tienda_frutas_colores::main_objects": "La tarea se completa cuando el alumno, en español, menciona las manzanas, las peras y las naranjas. Se aceptan errores gramaticales menores si el significado es claro.",
  "tienda_frutas_colores::actions": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores de frutas: manzanas rojas, peras verdes o naranjas naranjas. Se aceptan errores gramaticales menores si el significado es claro.",
  "tienda_frutas_colores::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos una cantidad de frutas: tres manzanas, cuatro peras o cinco naranjas. Se aceptan errores gramaticales menores si el significado es claro.",
  "tienda_frutas_colores::layout": "La tarea se completa cuando el alumno, en español, menciona al vendedor detrás de las cestas. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_sol_arbol_banco_perro::main_objects": "La tarea se completa cuando el alumno, en español, menciona el árbol, el banco y el perro. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_sol_arbol_banco_perro::actions": "La tarea se completa cuando el alumno, en español, menciona que el parque está soleado. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_sol_arbol_banco_perro::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores: árbol o hierba verde, banco o perro marrón, cielo azul. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_sol_arbol_banco_perro::layout": "La tarea se completa cuando el alumno, en español, menciona dónde está el perro: cerca del banco o sobre la hierba. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_cocina_cena::main_objects": "La tarea se completa cuando el alumno, en español, menciona a los miembros de la familia en la cocina. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_cocina_cena::actions": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones: cortar tomates, remover la sopa o lavar lechuga. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_cocina_cena::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona los alimentos: tomates, sopa, lechuga, pan o platos. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_cocina_cena::layout": "La tarea se completa cuando el alumno, en español, menciona la luz de la cocina por la tarde. Se aceptan errores gramaticales menores si el significado es claro.",
  "alumno_mochila_manana::main_objects": "La tarea se completa cuando el alumno, en español, menciona al estudiante y la mochila. Se aceptan errores gramaticales menores si el significado es claro.",
  "alumno_mochila_manana::actions": "La tarea se completa cuando el alumno, en español, menciona que es por la mañana o que entra sol por la ventana. Se aceptan errores gramaticales menores si el significado es claro.",
  "alumno_mochila_manana::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos tres objetos escolares: cuaderno, lápiz, libro, fiambrera o mochila. Se aceptan errores gramaticales menores si el significado es claro.",
  "alumno_mochila_manana::layout": "La tarea se completa cuando el alumno, en español, menciona que el cuaderno se está metiendo en la mochila. Se aceptan errores gramaticales menores si el significado es claro.",
  "cafe_dos_visitantes::main_objects": "La tarea se completa cuando el alumno, en español, menciona la cafetería y dos visitantes. Se aceptan errores gramaticales menores si el significado es claro.",
  "cafe_dos_visitantes::actions": "La tarea se completa cuando el alumno, en español, menciona las bebidas: café y zumo de naranja. Se aceptan errores gramaticales menores si el significado es claro.",
  "cafe_dos_visitantes::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al camarero y el menú. Se aceptan errores gramaticales menores si el significado es claro.",
  "cafe_dos_visitantes::layout": "La tarea se completa cuando el alumno, en español, menciona las sillas o la planta verde junto a la ventana. Se aceptan errores gramaticales menores si el significado es claro.",
  "persona_compra_boleto_taquilla::main_objects": "La tarea se completa cuando el alumno, en español, menciona al cliente y al cajero. Se aceptan errores gramaticales menores si el significado es claro.",
  "persona_compra_boleto_taquilla::actions": "La tarea se completa cuando el alumno, en español, menciona que la persona compra o recibe un billete. Se aceptan errores gramaticales menores si el significado es claro.",
  "persona_compra_boleto_taquilla::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona el dinero y el ventanillo de cristal. Se aceptan errores gramaticales menores si el significado es claro.",
  "persona_compra_boleto_taquilla::layout": "La tarea se completa cuando el alumno, en español, menciona el cartel azul encima del ventanillo. Se aceptan errores gramaticales menores si el significado es claro.",
  "patio_ninos_juegan_pelota::main_objects": "La tarea se completa cuando el alumno, en español, menciona el patio y los niños. Se aceptan errores gramaticales menores si el significado es claro.",
  "patio_ninos_juegan_pelota::actions": "La tarea se completa cuando el alumno, en español, menciona que hay tres niños. Se aceptan errores gramaticales menores si el significado es claro.",
  "patio_ninos_juegan_pelota::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones: patear, correr o esperar. Se aceptan errores gramaticales menores si el significado es claro.",
  "patio_ninos_juegan_pelota::layout": "La tarea se completa cuando el alumno, en español, menciona la pelota roja o el banco gris. Se aceptan errores gramaticales menores si el significado es claro.",
  "playa_sombrilla_toalla_mar::main_objects": "La tarea se completa cuando el alumno, en español, menciona la playa, el mar, la sombrilla y la toalla. Se aceptan errores gramaticales menores si el significado es claro.",
  "playa_sombrilla_toalla_mar::actions": "La tarea se completa cuando el alumno, en español, menciona que hace sol o que es una escena de playa. Se aceptan errores gramaticales menores si el significado es claro.",
  "playa_sombrilla_toalla_mar::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos colores: mar azul, sombrilla amarilla, toalla roja o sandalias verdes. Se aceptan errores gramaticales menores si el significado es claro.",
  "playa_sombrilla_toalla_mar::layout": "La tarea se completa cuando el alumno, en español, menciona las sandalias junto a la toalla. Se aceptan errores gramaticales menores si el significado es claro.",
  "medico_paciente_consulta::main_objects": "La tarea se completa cuando el alumno, en español, menciona al médico y al paciente. Se aceptan errores gramaticales menores si el significado es claro.",
  "medico_paciente_consulta::actions": "La tarea se completa cuando el alumno, en español, menciona que el médico usa un estetoscopio. Se aceptan errores gramaticales menores si el significado es claro.",
  "medico_paciente_consulta::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona la consulta o el entorno médico. Se aceptan errores gramaticales menores si el significado es claro.",
  "medico_paciente_consulta::layout": "La tarea se completa cuando el alumno, en español, menciona el ordenador o la carpeta azul en el escritorio. Se aceptan errores gramaticales menores si el significado es claro.",
  "turista_mapa_metro::main_objects": "La tarea se completa cuando el alumno, en español, menciona al turista y el mapa. Se aceptan errores gramaticales menores si el significado es claro.",
  "turista_mapa_metro::actions": "La tarea se completa cuando el alumno, en español, menciona la entrada del metro o el cartel del metro. Se aceptan errores gramaticales menores si el significado es claro.",
  "turista_mapa_metro::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona que el turista mira el mapa. Se aceptan errores gramaticales menores si el significado es claro.",
  "turista_mapa_metro::layout": "La tarea se completa cuando el alumno, en español, menciona los edificios o el paso de peatones detrás del turista. Se aceptan errores gramaticales menores si el significado es claro.",
  "picnic_lago_perro_pajaros::main_objects": "La tarea se completa cuando el alumno, en español, menciona el picnic junto al lago. Se aceptan errores gramaticales menores si el significado es claro.",
  "picnic_lago_perro_pajaros::actions": "La tarea se completa cuando el alumno, en español, menciona comida o bebida en la manta: bocadillos, manzanas, cesta o agua. Se aceptan errores gramaticales menores si el significado es claro.",
  "picnic_lago_perro_pajaros::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona el perro y los pájaros. Se aceptan errores gramaticales menores si el significado es claro.",
  "picnic_lago_perro_pajaros::layout": "La tarea se completa cuando el alumno, en español, menciona al menos una acción: sentarse, comer o servir agua. Se aceptan errores gramaticales menores si el significado es claro.",
  "mercado_manana_verduras::main_objects": "La tarea se completa cuando el alumno, en español, menciona el mercado matutino con vendedores y clientes. Se aceptan errores gramaticales menores si el significado es claro.",
  "mercado_manana_verduras::actions": "La tarea se completa cuando el alumno, en español, menciona al menos tres verduras: tomates, pepinos, zanahorias o patatas. Se aceptan errores gramaticales menores si el significado es claro.",
  "mercado_manana_verduras::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona las tarjetas de precio o los números visibles. Se aceptan errores gramaticales menores si el significado es claro.",
  "mercado_manana_verduras::layout": "La tarea se completa cuando el alumno, en español, menciona al cliente con una bolsa de tela. Se aceptan errores gramaticales menores si el significado es claro.",
  "apartamento_mudanza_cajas::main_objects": "La tarea se completa cuando el alumno, en español, menciona el apartamento y las cajas de mudanza. Se aceptan errores gramaticales menores si el significado es claro.",
  "apartamento_mudanza_cajas::actions": "La tarea se completa cuando el alumno, en español, menciona el número o la ubicación de las cajas. Se aceptan errores gramaticales menores si el significado es claro.",
  "apartamento_mudanza_cajas::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos una acción de mudanza: cerrar una caja con cinta o llevar una silla. Se aceptan errores gramaticales menores si el significado es claro.",
  "apartamento_mudanza_cajas::layout": "La tarea se completa cuando el alumno, en español, menciona la mesa envuelta o la alfombra azul enrollada. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciclista_repara_rueda::main_objects": "La tarea se completa cuando el alumno, en español, menciona al ciclista y la bicicleta. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciclista_repara_rueda::actions": "La tarea se completa cuando el alumno, en español, menciona que el ciclista repara una rueda. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciclista_repara_rueda::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos herramientas: bomba, llave inglesa o kit de parches. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciclista_repara_rueda::layout": "La tarea se completa cuando el alumno, en español, menciona el casco rojo o la bicicleta negra. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_ropa_tienda::main_objects": "La tarea se completa cuando el alumno, en español, menciona a la familia en una tienda de ropa. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_ropa_tienda::actions": "La tarea se completa cuando el alumno, en español, menciona al menos tres prendas: chaqueta, jersey, gorro o camisas. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_ropa_tienda::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona los colores de la ropa. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_ropa_tienda::layout": "La tarea se completa cuando el alumno, en español, menciona las tallas o el espejo. Se aceptan errores gramaticales menores si el significado es claro.",
  "oficina_colegas_computadores::main_objects": "La tarea se completa cuando el alumno, en español, menciona la oficina y los compañeros. Se aceptan errores gramaticales menores si el significado es claro.",
  "oficina_colegas_computadores::actions": "La tarea se completa cuando el alumno, en español, menciona ordenadores o portátiles. Se aceptan errores gramaticales menores si el significado es claro.",
  "oficina_colegas_computadores::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones de trabajo: escribir, llamar con auriculares o tomar notas. Se aceptan errores gramaticales menores si el significado es claro.",
  "oficina_colegas_computadores::layout": "La tarea se completa cuando el alumno, en español, menciona la impresora o el reloj que marca las 10:00. Se aceptan errores gramaticales menores si el significado es claro.",
  "aeropuerto_maletas_tablero::main_objects": "La tarea se completa cuando el alumno, en español, menciona la sala del aeropuerto y los pasajeros. Se aceptan errores gramaticales menores si el significado es claro.",
  "aeropuerto_maletas_tablero::actions": "La tarea se completa cuando el alumno, en español, menciona maletas o equipaje con colores o cantidades. Se aceptan errores gramaticales menores si el significado es claro.",
  "aeropuerto_maletas_tablero::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona el tablero de salidas con horarios. Se aceptan errores gramaticales menores si el significado es claro.",
  "aeropuerto_maletas_tablero::layout": "La tarea se completa cuando el alumno, en español, menciona al pasajero que señala el tablero. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_invierno_nieve_muneco::main_objects": "La tarea se completa cuando el alumno, en español, menciona el parque invernal y la nieve. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_invierno_nieve_muneco::actions": "La tarea se completa cuando el alumno, en español, menciona a los dos niños y al muñeco de nieve. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_invierno_nieve_muneco::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos detalles del muñeco: nariz de zanahoria, bufanda azul o sombrero negro. Se aceptan errores gramaticales menores si el significado es claro.",
  "parque_invierno_nieve_muneco::layout": "La tarea se completa cuando el alumno, en español, menciona la nieve que cae o el estanque helado. Se aceptan errores gramaticales menores si el significado es claro.",
  "cola_policlinica_documentos::main_objects": "La tarea se completa cuando el alumno, en español, menciona la sala de espera de la clínica y la cola. Se aceptan errores gramaticales menores si el significado es claro.",
  "cola_policlinica_documentos::actions": "La tarea se completa cuando el alumno, en español, menciona que cinco personas esperan. Se aceptan errores gramaticales menores si el significado es claro.",
  "cola_policlinica_documentos::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona los documentos o la carpeta azul. Se aceptan errores gramaticales menores si el significado es claro.",
  "cola_policlinica_documentos::layout": "La tarea se completa cuando el alumno, en español, menciona a la enfermera, al niño cansado, al bastón o al cartel de la cruz roja. Se aceptan errores gramaticales menores si el significado es claro.",
  "estudiantes_proyecto_biblioteca::main_objects": "La tarea se completa cuando el alumno, en español, menciona a los estudiantes que discuten un proyecto en la biblioteca. Se aceptan errores gramaticales menores si el significado es claro.",
  "estudiantes_proyecto_biblioteca::actions": "La tarea se completa cuando el alumno, en español, describe al menos dos roles o acciones entre señalar, tomar notas, sostener libros o escuchar. Se aceptan errores gramaticales menores si el significado es claro.",
  "estudiantes_proyecto_biblioteca::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona el portátil y el cartel con notas adhesivas. Se aceptan errores gramaticales menores si el significado es claro.",
  "estudiantes_proyecto_biblioteca::layout": "La tarea se completa cuando el alumno, en español, menciona las estanterías o el ambiente tranquilo de la biblioteca. Se aceptan errores gramaticales menores si el significado es claro.",
  "vecinos_proteccion_agua::main_objects": "La tarea se completa cuando el alumno, en español, menciona la fuga de agua en el pasillo de un edificio. Se aceptan errores gramaticales menores si el significado es claro.",
  "vecinos_proteccion_agua::actions": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones de ayuda: sostener un cubo, fregar, llamar o usar toallas. Se aceptan errores gramaticales menores si el significado es claro.",
  "vecinos_proteccion_agua::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona la mancha húmeda del techo o el agua en el suelo. Se aceptan errores gramaticales menores si el significado es claro.",
  "vecinos_proteccion_agua::layout": "La tarea se completa cuando el alumno, en español, menciona el ambiente de preocupación pero de cooperación. Se aceptan errores gramaticales menores si el significado es claro.",
  "excursion_ciudad_antigua::main_objects": "La tarea se completa cuando el alumno, en español, menciona la visita guiada en una ciudad antigua. Se aceptan errores gramaticales menores si el significado es claro.",
  "excursion_ciudad_antigua::actions": "La tarea se completa cuando el alumno, en español, menciona al guía y a los seis turistas. Se aceptan errores gramaticales menores si el significado es claro.",
  "excursion_ciudad_antigua::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona la arquitectura antigua: arcos de piedra, torre del reloj o calle empedrada. Se aceptan errores gramaticales menores si el significado es claro.",
  "excursion_ciudad_antigua::layout": "La tarea se completa cuando el alumno, en español, menciona al turista que hace fotos o el folleto. Se aceptan errores gramaticales menores si el significado es claro.",
  "voluntarios_limpian_rio::main_objects": "La tarea se completa cuando el alumno, en español, menciona a los voluntarios que limpian la orilla del río. Se aceptan errores gramaticales menores si el significado es claro.",
  "voluntarios_limpian_rio::actions": "La tarea se completa cuando el alumno, en español, menciona al menos dos tipos de basura o contenedores: botellas de plástico, papel, latas, bolsa azul o caja verde. Se aceptan errores gramaticales menores si el significado es claro.",
  "voluntarios_limpian_rio::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones entre recoger, levantar, revisar o separar. Se aceptan errores gramaticales menores si el significado es claro.",
  "voluntarios_limpian_rio::layout": "La tarea se completa cuando el alumno, en español, menciona la mañana nublada, el río o los árboles. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_planifica_vacaciones::main_objects": "La tarea se completa cuando el alumno, en español, menciona a la familia que planifica unas vacaciones. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_planifica_vacaciones::actions": "La tarea se completa cuando el alumno, en español, menciona objetos de planificación: mapa, portátil, folletos, cuaderno o presupuesto. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_planifica_vacaciones::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos preferencias o destinos mostrados: playa y montañas. Se aceptan errores gramaticales menores si el significado es claro.",
  "familia_planifica_vacaciones::layout": "La tarea se completa cuando el alumno, en español, menciona quién señala, escribe, muestra o mira. Se aceptan errores gramaticales menores si el significado es claro.",
  "taller_reparacion_bicicleta::main_objects": "La tarea se completa cuando el alumno, en español, menciona el taller de bicicletas y la bicicleta roja. Se aceptan errores gramaticales menores si el significado es claro.",
  "taller_reparacion_bicicleta::actions": "La tarea se completa cuando el alumno, en español, menciona que el mecánico repara o aprieta la cadena. Se aceptan errores gramaticales menores si el significado es claro.",
  "taller_reparacion_bicicleta::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos herramientas o piezas: neumático de repuesto, aceite, llave inglesa, destornillador o cadena. Se aceptan errores gramaticales menores si el significado es claro.",
  "taller_reparacion_bicicleta::layout": "La tarea se completa cuando el alumno, en español, menciona el banco de trabajo o los cajones con piezas. Se aceptan errores gramaticales menores si el significado es claro.",
  "entrevista_trabajo_joven::main_objects": "La tarea se completa cuando el alumno, en español, menciona la entrevista de trabajo y los participantes. Se aceptan errores gramaticales menores si el significado es claro.",
  "entrevista_trabajo_joven::actions": "La tarea se completa cuando el alumno, en español, menciona el currículum y tomar notas o leer. Se aceptan errores gramaticales menores si el significado es claro.",
  "entrevista_trabajo_joven::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona las emociones del joven: nervioso pero educado. Se aceptan errores gramaticales menores si el significado es claro.",
  "entrevista_trabajo_joven::layout": "La tarea se completa cuando el alumno, en español, menciona los detalles de la mesa de la oficina, como agua o una planta. Se aceptan errores gramaticales menores si el significado es claro.",
  "repartidor_lluvia_paquete::main_objects": "La tarea se completa cuando el alumno, en español, menciona al repartidor que entrega un paquete. Se aceptan errores gramaticales menores si el significado es claro.",
  "repartidor_lluvia_paquete::actions": "La tarea se completa cuando el alumno, en español, menciona la lluvia intensa, los charcos o la calle mojada. Se aceptan errores gramaticales menores si el significado es claro.",
  "repartidor_lluvia_paquete::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al cliente bajo el portal con un paraguas. Se aceptan errores gramaticales menores si el significado es claro.",
  "repartidor_lluvia_paquete::layout": "La tarea se completa cuando el alumno, en español, menciona colores como el impermeable amarillo o la caja marrón. Se aceptan errores gramaticales menores si el significado es claro.",
  "estacion_noche_tren_retrasado::main_objects": "La tarea se completa cuando el alumno, en español, menciona la estación de tren por la tarde y la situación del tren retrasado. Se aceptan errores gramaticales menores si el significado es claro.",
  "estacion_noche_tren_retrasado::actions": "La tarea se completa cuando el alumno, en español, menciona el tablero que muestra las 21:40 y un símbolo de retraso. Se aceptan errores gramaticales menores si el significado es claro.",
  "estacion_noche_tren_retrasado::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona a los pasajeros que esperan con maletas. Se aceptan errores gramaticales menores si el significado es claro.",
  "estacion_noche_tren_retrasado::layout": "La tarea se completa cuando el alumno, en español, menciona al menos dos acciones: mirar el móvil, dormir, señalar o sentarse. Se aceptan errores gramaticales menores si el significado es claro.",
  "coworking_discusion_proyecto::main_objects": "La tarea se completa cuando el alumno, en español, menciona la discusión del proyecto en un coworking y el equipo. Se aceptan errores gramaticales menores si el significado es claro.",
  "coworking_discusion_proyecto::actions": "La tarea se completa cuando el alumno, en español, describe al menos tres roles o aportaciones: maqueta en tableta, código, prioridades o gráficos. Se aceptan errores gramaticales menores si el significado es claro.",
  "coworking_discusion_proyecto::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona la colaboración o el ambiente de concentración. Se aceptan errores gramaticales menores si el significado es claro.",
  "coworking_discusion_proyecto::layout": "La tarea se completa cuando el alumno, en español, menciona detalles de la mesa como tazas de café o cables de carga. Se aceptan errores gramaticales menores si el significado es claro.",
  "plaza_festival_ciudad::main_objects": "La tarea se completa cuando el alumno, en español, menciona el festival en una plaza de la ciudad. Se aceptan errores gramaticales menores si el significado es claro.",
  "plaza_festival_ciudad::actions": "La tarea se completa cuando el alumno, en español, menciona el escenario, los músicos, las banderas y la multitud. Se aceptan errores gramaticales menores si el significado es claro.",
  "plaza_festival_ciudad::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona al menos dos actividades: mirar, bailar o comprar snacks. Se aceptan errores gramaticales menores si el significado es claro.",
  "plaza_festival_ciudad::layout": "La tarea se completa cuando el alumno, en español, menciona la fuente o el ayuntamiento antiguo. Se aceptan errores gramaticales menores si el significado es claro.",
  "discusion_presupuesto_familia::main_objects": "La tarea se completa cuando el alumno, en español, menciona la discusión del presupuesto familiar en la cocina. Se aceptan errores gramaticales menores si el significado es claro.",
  "discusion_presupuesto_familia::actions": "La tarea se completa cuando el alumno, en español, menciona evidencias de planificación del dinero: billetes, calculadora o cuaderno con números. Se aceptan errores gramaticales menores si el significado es claro.",
  "discusion_presupuesto_familia::colors_numbers": "La tarea se completa cuando el alumno, en español, describe al menos dos emociones o detalles del lenguaje corporal: tensión, brazos cruzados, enfado o silencio. Se aceptan errores gramaticales menores si el significado es claro.",
  "discusion_presupuesto_familia::layout": "La tarea se completa cuando el alumno, en español, menciona el teléfono o el vaso de leche. Se aceptan errores gramaticales menores si el significado es claro.",
  "museo_arte_moderno_visitantes::main_objects": "La tarea se completa cuando el alumno, en español, menciona la galería de un museo de arte moderno. Se aceptan errores gramaticales menores si el significado es claro.",
  "museo_arte_moderno_visitantes::actions": "La tarea se completa cuando el alumno, en español, describe el cuadro abstracto con formas y colores. Se aceptan errores gramaticales menores si el significado es claro.",
  "museo_arte_moderno_visitantes::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona distintas reacciones de los visitantes: sonrisa, confusión o tomar notas. Se aceptan errores gramaticales menores si el significado es claro.",
  "museo_arte_moderno_visitantes::layout": "La tarea se completa cuando el alumno, en español, menciona la escultura en forma de cubo plateado. Se aceptan errores gramaticales menores si el significado es claro.",
  "rueda_prensa_deporte::main_objects": "La tarea se completa cuando el alumno, en español, menciona la rueda de prensa después del partido. Se aceptan errores gramaticales menores si el significado es claro.",
  "rueda_prensa_deporte::actions": "La tarea se completa cuando el alumno, en español, menciona al entrenador, al jugador, a los periodistas y a los micrófonos. Se aceptan errores gramaticales menores si el significado es claro.",
  "rueda_prensa_deporte::colors_numbers": "La tarea se completa cuando el alumno, en español, describe al menos dos acciones de los medios: cámaras, grabación de audio o mano levantada para preguntar. Se aceptan errores gramaticales menores si el significado es claro.",
  "rueda_prensa_deporte::layout": "La tarea se completa cuando el alumno, en español, menciona el marcador 2-1 o las emociones del entrenador y del jugador. Se aceptan errores gramaticales menores si el significado es claro.",
  "reunion_vecinos_parque_nuevo::main_objects": "La tarea se completa cuando el alumno, en español, menciona la reunión vecinal sobre un parque nuevo. Se aceptan errores gramaticales menores si el significado es claro.",
  "reunion_vecinos_parque_nuevo::actions": "La tarea se completa cuando el alumno, en español, menciona al urbanista y el mapa con árboles, caminos y zona de juegos. Se aceptan errores gramaticales menores si el significado es claro.",
  "reunion_vecinos_parque_nuevo::colors_numbers": "La tarea se completa cuando el alumno, en español, describe las reacciones o la participación mixta de los vecinos. Se aceptan errores gramaticales menores si el significado es claro.",
  "reunion_vecinos_parque_nuevo::layout": "La tarea se completa cuando el alumno, en español, menciona el cartel de antes y después. Se aceptan errores gramaticales menores si el significado es claro.",
  "startup_app_inversor::main_objects": "La tarea se completa cuando el alumno, en español, menciona la presentación de la startup a un inversor. Se aceptan errores gramaticales menores si el significado es claro.",
  "startup_app_inversor::actions": "La tarea se completa cuando el alumno, en español, menciona el panel de la app, el portátil y el teléfono prototipo. Se aceptan errores gramaticales menores si el significado es claro.",
  "startup_app_inversor::colors_numbers": "La tarea se completa cuando el alumno, en español, describe los roles: el que presenta, quien controla el portátil o el inversor con cuaderno. Se aceptan errores gramaticales menores si el significado es claro.",
  "startup_app_inversor::layout": "La tarea se completa cuando el alumno, en español, menciona la expresión seria del inversor. Se aceptan errores gramaticales menores si el significado es claro.",
  "calle_despues_lluvia_reparacion_trafico::main_objects": "La tarea se completa cuando el alumno, en español, menciona la calle después de una lluvia intensa. Se aceptan errores gramaticales menores si el significado es claro.",
  "calle_despues_lluvia_reparacion_trafico::actions": "La tarea se completa cuando el alumno, en español, menciona el tráfico: coches esperando y autobús parado. Se aceptan errores gramaticales menores si el significado es claro.",
  "calle_despues_lluvia_reparacion_trafico::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona a los trabajadores que reparan un desagüe bloqueado. Se aceptan errores gramaticales menores si el significado es claro.",
  "calle_despues_lluvia_reparacion_trafico::layout": "La tarea se completa cuando el alumno, en español, menciona consecuencias como un charco, hojas mojadas o ramas rotas. Se aceptan errores gramaticales menores si el significado es claro.",
  "juicio_abogados_testigo::main_objects": "La tarea se completa cuando el alumno, en español, menciona la audiencia formal en un tribunal. Se aceptan errores gramaticales menores si el significado es claro.",
  "juicio_abogados_testigo::actions": "La tarea se completa cuando el alumno, en español, identifica los roles: juez, abogados, testigo, secretario u observadores. Se aceptan errores gramaticales menores si el significado es claro.",
  "juicio_abogados_testigo::colors_numbers": "La tarea se completa cuando el alumno, en español, describe acciones procesales: presentar un documento, revisar notas, escribir o testigo ante el micrófono. Se aceptan errores gramaticales menores si el significado es claro.",
  "juicio_abogados_testigo::layout": "La tarea se completa cuando el alumno, en español, menciona la disposición formal de los asientos o los observadores en silencio. Se aceptan errores gramaticales menores si el significado es claro.",
  "debate_politico_estudio_tv::main_objects": "La tarea se completa cuando el alumno, en español, menciona el debate político en televisión. Se aceptan errores gramaticales menores si el significado es claro.",
  "debate_politico_estudio_tv::actions": "La tarea se completa cuando el alumno, en español, identifica a los candidatos, al moderador, a los podios, a las cámaras y al público. Se aceptan errores gramaticales menores si el significado es claro.",
  "debate_politico_estudio_tv::colors_numbers": "La tarea se completa cuando el alumno, en español, contrasta el lenguaje corporal o las expresiones de los candidatos. Se aceptan errores gramaticales menores si el significado es claro.",
  "debate_politico_estudio_tv::layout": "La tarea se completa cuando el alumno, en español, menciona la pantalla roja y azul sin eslóganes. Se aceptan errores gramaticales menores si el significado es claro.",
  "conferencia_cientifica_poster::main_objects": "La tarea se completa cuando el alumno, en español, menciona la sesión de pósters de una conferencia científica. Se aceptan errores gramaticales menores si el significado es claro.",
  "conferencia_cientifica_poster::actions": "La tarea se completa cuando el alumno, en español, identifica al investigador y a los tres asistentes. Se aceptan errores gramaticales menores si el significado es claro.",
  "conferencia_cientifica_poster::colors_numbers": "La tarea se completa cuando el alumno, en español, describe materiales abstractos: diagramas, gráficos, dibujo molecular o gráfica. Se aceptan errores gramaticales menores si el significado es claro.",
  "conferencia_cientifica_poster::layout": "La tarea se completa cuando el alumno, en español, menciona preguntas, señalar, notas en tableta, credenciales o tazas de café. Se aceptan errores gramaticales menores si el significado es claro.",
  "negociacion_contrato_sala::main_objects": "La tarea se completa cuando el alumno, en español, menciona la negociación de un contrato en una sala de reuniones. Se aceptan errores gramaticales menores si el significado es claro.",
  "negociacion_contrato_sala::actions": "La tarea se completa cuando el alumno, en español, identifica a los dos equipos y los materiales de negociación: contratos, hoja de cálculo o bolígrafos. Se aceptan errores gramaticales menores si el significado es claro.",
  "negociacion_contrato_sala::colors_numbers": "La tarea se completa cuando el alumno, en español, describe acciones estratégicas: subrayar una cláusula, susurrar o ofrecer un apretón de manos no aceptado. Se aceptan errores gramaticales menores si el significado es claro.",
  "negociacion_contrato_sala::layout": "La tarea se completa cuando el alumno, en español, menciona el ambiente de cautela. Se aceptan errores gramaticales menores si el significado es claro.",
  "redaccion_noticias_urgente::main_objects": "La tarea se completa cuando el alumno, en español, menciona la redacción de noticias de última hora. Se aceptan errores gramaticales menores si el significado es claro.",
  "redaccion_noticias_urgente::actions": "La tarea se completa cuando el alumno, en español, identifica roles y acciones: editores, productor, presentador, ordenadores o micrófono. Se aceptan errores gramaticales menores si el significado es claro.",
  "redaccion_noticias_urgente::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona señales visuales de urgencia: pantalla en la pared con mapa e iconos de alerta, reloj 18:05. Se aceptan errores gramaticales menores si el significado es claro.",
  "redaccion_noticias_urgente::layout": "La tarea se completa cuando el alumno, en español, describe la presión y la concentración, además del desorden en los escritorios. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciudad_futura_transporte_energia_ecologia::main_objects": "La tarea se completa cuando el alumno, en español, menciona la ciudad del futuro y su transporte limpio. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciudad_futura_transporte_energia_ecologia::actions": "La tarea se completa cuando el alumno, en español, menciona al menos dos tipos de transporte: autobuses eléctricos, pods autónomos o ciclistas. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciudad_futura_transporte_energia_ecologia::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona elementos de energía o ecología: paneles solares, turbinas eólicas, jardines verticales, árboles o aire limpio. Se aceptan errores gramaticales menores si el significado es claro.",
  "ciudad_futura_transporte_energia_ecologia::layout": "La tarea se completa cuando el alumno, en español, hace una comparación o hipótesis basada solo en el diseño sostenible visible. Se aceptan errores gramaticales menores si el significado es claro.",
  "restauradores_cuadro_danado::main_objects": "La tarea se completa cuando el alumno, en español, menciona el taller de restauración y el cuadro dañado. Se aceptan errores gramaticales menores si el significado es claro.",
  "restauradores_cuadro_danado::actions": "La tarea se completa cuando el alumno, en español, describe los daños visibles: retrato descolorido, esquina rota o grietas oscuras. Se aceptan errores gramaticales menores si el significado es claro.",
  "restauradores_cuadro_danado::colors_numbers": "La tarea se completa cuando el alumno, en español, menciona el proceso de restauración: pincel diminuto, lámpara de aumento o documentación en tableta. Se aceptan errores gramaticales menores si el significado es claro.",
  "restauradores_cuadro_danado::layout": "La tarea se completa cuando el alumno, en español, menciona herramientas o materiales como bastoncillos, tarros de pigmento, guantes o carta de colores. Se aceptan errores gramaticales menores si el significado es claro.",
  "encuentro_internacional_clima::main_objects": "La tarea se completa cuando el alumno, en español, menciona el encuentro internacional sobre el clima. Se aceptan errores gramaticales menores si el significado es claro.",
  "encuentro_internacional_clima::actions": "La tarea se completa cuando el alumno, en español, identifica a los delegados, al orador, a los traductores, la mesa circular y las banderas o placas con nombres. Se aceptan errores gramaticales menores si el significado es claro.",
  "encuentro_internacional_clima::colors_numbers": "La tarea se completa cuando el alumno, en español, describe evidencias climáticas en la pantalla: mapa del mundo, línea de temperatura ascendente o icono de inundación. Se aceptan errores gramaticales menores si el significado es claro.",
  "encuentro_internacional_clima::layout": "La tarea se completa cuando el alumno, en español, menciona notas, conversación en voz baja o cabinas de traducción. Se aceptan errores gramaticales menores si el significado es claro."
}


def apply_criteria_fixes(quest: dict) -> None:
    """Rotate misaligned English criteria for known quests (for audit only)."""
    code = quest["code"]
    fixes = CRITERIA_FIXES.get(code)
    if not fixes:
        return
    for task in quest["tasks"]:
        task_code = task["code"]
        if task_code in fixes:
            task["completion_criteria"] = fixes[task_code]


def build_es_quest(quest: dict) -> dict:
    out = dict(quest)
    out["course_code"] = "es_ru"
    tasks = []
    for task in quest["tasks"]:
        key = f"{quest['code']}::{task['code']}"
        criteria = SPANISH_CRITERIA.get(key)
        if criteria is None:
            raise KeyError(f"Missing Spanish criteria for {key}")
        tasks.append({
            "code": task["code"],
            "title": task["title"],
            "completion_criteria": criteria,
            "is_required": task["is_required"],
            "sort_order": task["sort_order"],
        })
    out["tasks"] = tasks
    return out


def build_en_quest(quest: dict) -> dict:
    fixed = json.loads(json.dumps(quest))
    apply_criteria_fixes(fixed)
    out = dict(fixed)
    out["course_code"] = "en_ru"
    tasks = []
    for task in fixed["tasks"]:
        criteria = task["completion_criteria"].replace("in Spanish", "in English")
        tasks.append({
            "code": task["code"],
            "title": task["title"],
            "completion_criteria": criteria,
            "is_required": task["is_required"],
            "sort_order": task["sort_order"],
        })
    out["tasks"] = tasks
    return out


def verify_russian_titles(quests: list[dict]) -> None:
    for quest in quests:
        if not quest.get("title") or not any("\u0400" <= ch <= "\u04FF" for ch in quest["title"]):
            raise ValueError(f"Quest {quest['code']} title is not Russian: {quest.get('title')!r}")
        for task in quest["tasks"]:
            if not task.get("title") or not any("\u0400" <= ch <= "\u04FF" for ch in task["title"]):
                raise ValueError(f"Quest {quest['code']} task {task['code']} title is not Russian")


def verify_es_output(quests: list[dict]) -> None:
    if len(quests) != 50:
        raise ValueError(f"Expected 50 quests, got {len(quests)}")
    verify_russian_titles(quests)
    for quest in quests:
        if quest.get("course_code") != "es_ru":
            raise ValueError(f"Quest {quest['code']} missing course_code es_ru")
        for task in quest["tasks"]:
            criteria = task["completion_criteria"]
            if not criteria.startswith("La tarea se completa cuando el alumno, en español,"):
                raise ValueError(f"Quest {quest['code']} task {task['code']} criteria not Spanish prefix")
            if not criteria.endswith("Se aceptan errores gramaticales menores si el significado es claro."):
                raise ValueError(f"Quest {quest['code']} task {task['code']} criteria missing Spanish suffix")
    print(f"Verified es_ru: {len(quests)} quests, {sum(len(q['tasks']) for q in quests)} tasks")


def verify_en_output(quests: list[dict]) -> None:
    if len(quests) != 50:
        raise ValueError(f"Expected 50 quests, got {len(quests)}")
    verify_russian_titles(quests)
    for quest in quests:
        if quest.get("course_code") != "en_ru":
            raise ValueError(f"Quest {quest['code']} missing course_code en_ru")
        for task in quest["tasks"]:
            criteria = task["completion_criteria"]
            if "in Spanish" in criteria:
                raise ValueError(f"Quest {quest['code']} task {task['code']} still references Spanish")
            if not criteria.startswith("Task is complete when the learner, in English,"):
                raise ValueError(f"Quest {quest['code']} task {task['code']} criteria not English prefix")
    print(f"Verified en_ru: {len(quests)} quests, {sum(len(q['tasks']) for q in quests)} tasks")


def load_source() -> list[dict]:
    if INPUT_PATH.exists():
        with INPUT_PATH.open(encoding="utf-8") as f:
            return json.load(f)
    raise FileNotFoundError(
        f"Input not found: {INPUT_PATH}. "
        "Export the English-criteria source with:\n"
        "  git show 0efe4a43:resources/picture_quests/es_ru_picture_quests_50.json > /tmp/orig_pq.json"
    )


def main() -> int:
    try:
        source = load_source()
    except FileNotFoundError as exc:
        print(str(exc), file=sys.stderr)
        return 1

    built_es = [build_es_quest(q) for q in source]
    built_en = [build_en_quest(q) for q in source]
    verify_es_output(built_es)
    verify_en_output(built_en)

    OUTPUT_ES_PATH.parent.mkdir(parents=True, exist_ok=True)
    with OUTPUT_ES_PATH.open("w", encoding="utf-8") as f:
        json.dump(built_es, f, ensure_ascii=False, indent=2)
        f.write("\n")
    with OUTPUT_EN_PATH.open("w", encoding="utf-8") as f:
        json.dump(built_en, f, ensure_ascii=False, indent=2)
        f.write("\n")

    print(f"Wrote {OUTPUT_ES_PATH}")
    print(f"Wrote {OUTPUT_EN_PATH}")
    fixed_codes = sorted(CRITERIA_FIXES.keys())
    print(f"Criteria alignment fixes applied for: {', '.join(fixed_codes)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

