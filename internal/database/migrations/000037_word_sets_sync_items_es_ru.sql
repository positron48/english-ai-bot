-- Sync es_ru word-set items from the current resources/wordsets/es_ru_seed_word_sets.json.
-- Supersedes 000035: re-applies the full current word list (not just the partial
-- snapshot taken at 000035 time) for every set that has words, and fixes the
-- cross-course tagging bug from 000035's ON CONFLICT clause: that migration kept
-- whatever course_code a colliding word_cards row already had (COALESCE preserved
-- it), so any of these words that happened to already exist (e.g. created untagged
-- by the word-set importer before it correctly propagated course_code, then picked
-- up by the English course's course-blind TTS prefetch) stayed wrongly bound to
-- en_ru. This migration forces course_code='es_ru' on conflict instead.
-- Run cmd/fix_misassigned_words beforehand to delete any word_cards already
-- corrupted that way (along with their stray en_ru TTS audio) so they get a clean
-- course_code='es_ru' row recreated here.

INSERT INTO word_cards (word, definition, course_code, updated_at)
VALUES
  ('yo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tú', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usted', '', 'es_ru', CURRENT_TIMESTAMP),
  ('él', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ella', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nosotros', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vosotros', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ustedes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('este', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ese', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aquel', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mi', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tu', '', 'es_ru', CURRENT_TIMESTAMP),
  ('su', '', 'es_ru', CURRENT_TIMESTAMP),
  ('qué', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quién', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuál', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuánto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cómo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuándo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dónde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adónde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('por', '', 'es_ru', CURRENT_TIMESTAMP),
  ('de', '', 'es_ru', CURRENT_TIMESTAMP),
  ('en', '', 'es_ru', CURRENT_TIMESTAMP),
  ('con', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sin', '', 'es_ru', CURRENT_TIMESTAMP),
  ('para', '', 'es_ru', CURRENT_TIMESTAMP),
  ('y', '', 'es_ru', CURRENT_TIMESTAMP),
  ('o', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sí', '', 'es_ru', CURRENT_TIMESTAMP),
  ('no', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muy', '', 'es_ru', CURRENT_TIMESTAMP),
  ('también', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tampoco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hola', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adiós', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gracias', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perdón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disculpa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bienvenido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saludo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('chao', '', 'es_ru', CURRENT_TIMESTAMP),
  ('familia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('madre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('padre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hijo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hija', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hermano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hermana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abuelo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('país', '', 'es_ru', CURRENT_TIMESTAMP),
  ('idioma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nacionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('español', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extranjero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persona', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hombre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mujer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ser', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tener', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hacer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('haber', '', 'es_ru', CURRENT_TIMESTAMP),
  ('existir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quedar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llamarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('venir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caminar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bajar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abrir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cerrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('querer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('poder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saber', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('necesitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deber', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puerta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mesa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('silla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('baño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bolso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mochila', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llave', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dinero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasaporte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarjeta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teléfono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escuela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clase', '', 'es_ru', CURRENT_TIMESTAMP),
  ('internet', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mensaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agua', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pan', '', 'es_ru', CURRENT_TIMESTAMP),
  ('leche', '', 'es_ru', CURRENT_TIMESTAMP),
  ('café', '', 'es_ru', CURRENT_TIMESTAMP),
  ('té', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carne', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fruta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verdura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calle', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plaza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tienda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('banco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hotel', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restaurante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autobús', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tren', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taxi', '', 'es_ru', CURRENT_TIMESTAMP),
  ('billete', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pagar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('uno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tres', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuatro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cinco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siete', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nueve', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cien', '', 'es_ru', CURRENT_TIMESTAMP),
  ('día', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mañana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('noche', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ahora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('blanco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rojo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('azul', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grande', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pequeño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nuevo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('me', '', 'es_ru', CURRENT_TIMESTAMP),
  ('te', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('la', '', 'es_ru', CURRENT_TIMESTAMP),
  ('le', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('os', '', 'es_ru', CURRENT_TIMESTAMP),
  ('los', '', 'es_ru', CURRENT_TIMESTAMP),
  ('las', '', 'es_ru', CURRENT_TIMESTAMP),
  ('les', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mí', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ti', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tuyo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suyo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nuestro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vuestro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('algo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alguien', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nadie', '', 'es_ru', CURRENT_TIMESTAMP),
  ('todo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('otro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hacia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hasta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('durante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('después', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detrás', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cerca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('porque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuando', '', 'es_ru', CURRENT_TIMESTAMP),
  ('si', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aunque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mientras', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entonces', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pues', '', 'es_ru', CURRENT_TIMESTAMP),
  ('además', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siempre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nunca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aún', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ya', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pronto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('luego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('primero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bastante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demasiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despertar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('levantarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ducharse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lavarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vestirse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desayunar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('almorzar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dormir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descansar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trabajar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estudiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cocinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('beber', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pedir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('servir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elegir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mercado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viajar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conducir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llegar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volver', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cruzar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esperar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correr', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tomar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recibir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ayudar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('buscar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encontrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llevar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dejar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('traer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sentir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pensar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('creer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recordar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('olvidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gustar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preferir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nombre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apellido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dirección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('número', '', 'es_ru', CURRENT_TIMESTAMP),
  ('firma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formulario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esposo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esposa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pareja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nieto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nieta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('primo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('joven', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('piso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apartamento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cocina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dormitorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sofá', '', 'es_ru', CURRENT_TIMESTAMP),
  ('armario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estantería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ducha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pared', '', 'es_ru', CURRENT_TIMESTAMP),
  ('techo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ropa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('camisa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pantalón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zapato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abrigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vestido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bajo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carné', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permiso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('copia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arroz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('queso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huevo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pollo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pescado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sopa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ensalada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manzana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plátano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('naranja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('azúcar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aceite', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zumo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('menú', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('camarero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reserva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('postre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bebida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vaso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cliente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cajero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descuento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oferta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recibo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('efectivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moneda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devolver', '', 'es_ru', CURRENT_TIMESTAMP),
  ('farmacia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medicina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('envío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paquete', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sello', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sucursal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('servicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('botella', '', 'es_ru', CURRENT_TIMESTAMP),
  ('envase', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bolsa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cartón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lata', '', 'es_ru', CURRENT_TIMESTAMP),
  ('kilo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('litro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gramo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trozo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('docena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('museo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cine', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teatro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospital', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oficina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aeropuerto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puerto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biblioteca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iglesia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comisaría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tranvía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('andén', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conductor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasajero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('horario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ruta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('izquierda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derecha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esquina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mapa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('camino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orientación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('norte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sur', '', 'es_ru', CURRENT_TIMESTAMP),
  ('centro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alojamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recepción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huésped', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alquilar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alquiler', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llavero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maleta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equipaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vuelta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('destino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lunes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('martes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('miércoles', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jueves', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viernes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sábado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('domingo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('febrero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marzo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invierno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('minuto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segundo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reloj', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calendario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fecha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cita', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reunión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('turno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puntual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semanal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mensual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frecuente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('raro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('duración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tiempo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sol', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lluvia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nieve', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nube', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('veinte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('treinta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuarenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cincuenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sesenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('setenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('último', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simpático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antipático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('serio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tranquilo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nervioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('feliz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('triste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cansado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocupado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('libre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('listo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viejo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('feo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limpio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sucio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fácil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('difícil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rápido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tristeza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('miedo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dolor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hambre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sed', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sueño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfermo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gusto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interés', '', 'es_ru', CURRENT_TIMESTAMP),
  ('favorito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dulce', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amargo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('igual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diferente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mejor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mayor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('menor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('largo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ancho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estrecho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lleno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fuerte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('débil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('importante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conmigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cualquiera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ninguno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('varios', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suficiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quienquiera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semejante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bastantes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sendos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respectivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('según', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excepto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salvo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incluido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incluso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('debido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('causa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fuera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dentro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alrededor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('junto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encima', '', 'es_ru', CURRENT_TIMESTAMP),
  ('debajo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tras', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('final', '', 'es_ru', CURRENT_TIMESTAMP),
  ('principio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siquiera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apenas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aparte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('todavía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('así', '', 'es_ru', CURRENT_TIMESTAMP),
  ('total', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enseguida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('finalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posteriormente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quizá', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quizás', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('claramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lentamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rápidamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fácilmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('difícilmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casi', '', 'es_ru', CURRENT_TIMESTAMP),
  ('raramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exactamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aproximadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('principalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('claro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vale', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bueno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vaya', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perdone', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oiga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mire', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repita', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perdona', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disculpe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entiendo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correcto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('efectivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('genial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acuerdo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('duda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limpiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barrer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fregar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lavar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planchar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ordenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arreglar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cambiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enchufe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bombilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escoba', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cubo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jabón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('toalla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sábana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('herramienta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('chaqueta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('falda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jersey', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calcetín', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gorra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cinturón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('talla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('color', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('champú', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cepillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peine', '', 'es_ru', CURRENT_TIMESTAMP),
  ('espejo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfume', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detergente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alfombra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('factura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarifa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gasto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depósito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inquilino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propietario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electricidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suministro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calefacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conexión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mensualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('importe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deuda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vencimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solicitud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('registro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inscripción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expediente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('archivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fotografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('identificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resguardo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impreso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respuesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pregunta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('página', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trámite', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ruido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fuga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mancha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rotura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urgente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('molestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avisar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llamar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solucionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prohibir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('queja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('favor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ayuda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vecino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('médico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('profesor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingeniero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abogado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfermero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cocinero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dependiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('director', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secretario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empleado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jefe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compañero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('técnico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('artista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periodista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('chofer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recepcionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limpiador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estudiante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trabajador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('organizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preparar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enviar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revisar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('firmar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imprimir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('copiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guardar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apuntar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anotar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confirmar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cancelar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reservar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('completar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entregar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recoger', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reunirse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colaborar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejercicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('examen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prueba', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nota', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pizarra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('libro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuaderno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarea', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proyecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('práctica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejemplo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corregir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enseñar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ordenador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pantalla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teclado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ratón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fichero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carpeta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aplicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraseña', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usuario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enlace', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descargar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sincronizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instalar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actualizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('borrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reiniciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transferir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conectar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desconectar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('localizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('navegar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('texto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('email', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llamada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('destinatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('chat', '', 'es_ru', CURRENT_TIMESTAMP),
  ('audio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vídeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('responder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marcar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colgar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sonar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grabar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjuntar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reenviar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('buzón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('señal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interfaz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sitio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('red', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acceso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('datos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('privacidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('buscador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('servidor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cabeza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cara', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ojo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oreja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nariz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('boca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuello', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hombro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brazo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dedo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pierna', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pie', '', 'es_ru', CURRENT_TIMESTAMP),
  ('espalda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estómago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiebre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gripe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resfriado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('molestia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cansancio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mareo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('náusea', '', 'es_ru', CURRENT_TIMESTAMP),
  ('herida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('golpe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alergia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('infección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfermedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('síntoma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vomitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('toser', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doler', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sangrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doctor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auxiliar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paciente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consulta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('receta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pastilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jarabe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inyección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacuna', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tratamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('análisis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('turnero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urgencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('póliza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('farmacéutico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('administrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deporte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fútbol', '', 'es_ru', CURRENT_TIMESTAMP),
  ('baloncesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('natación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciclismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gimnasio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equipo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('partido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jugar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ganar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('campeón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrenamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costumbre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hábito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dieta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saludable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('activo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relajado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estresado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fumar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reposar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respirar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moverse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mejorar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vehículo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coche', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bicicleta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('camión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carretera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autopista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tráfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semáforo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gasolina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aparcamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('garaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manejar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aparcar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrancar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vuelo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embarque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valija', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bulto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('control', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mostrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terminal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retraso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cancelación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('butaca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vagón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transbordo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llegada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adquirir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facturar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barrio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zona', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avenida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cruce', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rotonda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indicador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cartel', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desvío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bajada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('continuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('girar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reservación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuarto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conserje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('credencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visitante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pernocta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colchón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amenidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desayuno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limpieza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ascensor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balcón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrendar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mudarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disponible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reservado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ayuntamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('juzgado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colegio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('universidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supermercado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instituto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polideportivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('piscina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gasolinera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taller', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peluquería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('panadería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carnicería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verdulería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cafetería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discoteca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peligro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ladrón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('policía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emergencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('accidente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('daño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pérdida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraviar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('romper', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descuidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('denunciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recuperar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuidado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('opinión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('idea', '', 'es_ru', CURRENT_TIMESTAMP),
  ('razón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muestra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verdad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mentira', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interesante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aburrido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('útil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inútil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('necesario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imposible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('importar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aceptar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rechazar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dudar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ratificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acordar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discutir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('opinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convencido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inseguro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('talvez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depende', '', 'es_ru', CURRENT_TIMESTAMP),
  ('veraz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('falso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conformidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desacuerdo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carácter', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ánimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfadado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preocupado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sorprendido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asustado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orgulloso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tímido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tolerante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impaciente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cordial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('egoísta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amistad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respeto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visita', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invitado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anfitrión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grupo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conocido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compañía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conocer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relacionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saludar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despedirse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prometer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ofrecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concertar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('citarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apologizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agradecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perdonar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rogar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exigir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sugerir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insistir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retrasar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aplazar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resolver', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excusa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apoyo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('película', '', 'es_ru', CURRENT_TIMESTAMP),
  ('serie', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('música', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cantante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actriz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('realizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('novela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('poema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('boleto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hobby', '', 'es_ru', CURRENT_TIMESTAMP),
  ('afición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pintar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dibujar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bailar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cantar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tocar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('leer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escribir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fotografiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coleccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jardinería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('juego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('celebración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cumpleaños', '', 'es_ru', CURRENT_TIMESTAMP),
  ('boda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aniversario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regalo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sorpresa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brindis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decoración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pase', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agenda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asistir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('celebrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coordinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convocar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aceptación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('publicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comentario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('foto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clip', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plataforma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguidor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('noticia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('titular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compartir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('publicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('registrarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bloquear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('etiquetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('afiliarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('galería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monumento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('castillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palacio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediateca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sala', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auditorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escenario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festival', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingreso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recorrido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('naturaleza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('campo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bosque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('montaña', '', 'es_ru', CURRENT_TIMESTAMP),
  ('río', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('playa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('isla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valle', '', 'es_ru', CURRENT_TIMESTAMP),
  ('piedra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cielo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tierra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('animal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caballo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pájaro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('árbol', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hoja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hierba', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clima', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temperatura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tormenta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('niebla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hielo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llover', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nevar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soplar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nublado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soleado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('húmedo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('templado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingrediente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preparación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuchillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tenedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuchara', '', 'es_ru', CURRENT_TIMESTAMP),
  ('olla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sartén', '', 'es_ru', CURRENT_TIMESTAMP),
  ('horno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('freír', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hervir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mezclar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pelar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('añadir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degustar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sabor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('picante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caliente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fresco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asimismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inicialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conclusivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concretamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brevemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sinceramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fehacientemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('básicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esquemáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teóricamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empíricamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('analíticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consecuentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paralelamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('globalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esencialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciertamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recapitulando', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resumiendo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planteamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfoque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('argumento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apartado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('causalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('motivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consecuencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('efecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('finalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propósito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condicionante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('requisito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dependencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('determinante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detonante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incidencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repercusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derivación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desencadenante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provocar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocasionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facilitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impedir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contribuir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('justificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('determinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condicionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excepción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('objeción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salvedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restricción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matiz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('límite', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alternativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discrepancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('divergencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrapunto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrapeso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excluir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exceptuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distinguir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relativizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puntualizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certeza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incertidumbre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipótesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sospecha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evidencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conjetura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previsión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estimación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cálculo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convicción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacilación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plausible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verosímil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('improbable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presumible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dudoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anterioridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posterioridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simultaneidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sucesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precedencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intervalo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periodo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plazo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('continuidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interrupción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pausa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reanudación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cierre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transcurso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preceder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suceder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reanudarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prolongarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culminar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resumir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sintetizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aclarar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precisar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detallar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reformular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parafrasear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('citar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mencionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejemplificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ilustrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('señalar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('destacar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agregar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('omitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('narrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relatar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('describir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interpretar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apostillar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('referencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fuente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('versión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esclarecimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ilustración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antecedente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aclaración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('candidatura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('currículum', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrevista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('selección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('candidato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reclutador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disponibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recomendación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trayectoria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portafolio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('experiencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semblanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aptitud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contratación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jornada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('postular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contratar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ascender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incorporarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renunciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preselección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('responsabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('función', '', 'es_ru', CURRENT_TIMESTAMP),
  ('competencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('destreza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exigencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sueldo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('beneficio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puntualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iniciativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('liderazgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('productividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autonomía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flexibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compromiso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diligente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eficaz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('organizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cumplir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('armonizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dominar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfeccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supervisar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gestionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capacitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orden', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encuentro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunicado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aviso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propuesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interpelación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprobación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('petición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convocatoria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remitente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('receptor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solicitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trasladar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anexar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('archivar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plan', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cometido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cadencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrega', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prioridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avance', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recurso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presupuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('responsable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cronograma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('objetivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('etapa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bloqueo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cómputo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asignación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interdependencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pendiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asignar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('finalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprobar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lanzar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clausurar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fallo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conflicto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tensión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reclamación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crítica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incidente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fricción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('malentendido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuelas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disculparse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rectificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solventar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esclarecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asumir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plantear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teletrabajo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remoto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autónomo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('freelance', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encargo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('videollamada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portátil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colaboración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facturación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asincrónico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('briefing', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anticipo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entregable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('externalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subcontratar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coworking', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consultoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('honorario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proveedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autónomamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslocalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descentralizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subcontratación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intermediación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facturable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retainer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cotización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consultor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nomadismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asignatura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('materia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facultad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carrera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('campus', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semestre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aula', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alumno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('literatura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matemática', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('física', '', 'es_ru', CURRENT_TIMESTAMP),
  ('química', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filosofía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('economía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('psicología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sociología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprobado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suspenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diploma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matrícula', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corrección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recuperación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repasar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprobar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suspender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('baremo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rúbrica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tutoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simulacro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escrito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprendizaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conocimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('método', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memoria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concentración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disciplina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curiosidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esfuerzo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('progreso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estrategia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rutina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asimilar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejercitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capacitarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfeccionarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reforzar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autodidacta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lectura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('investigación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('artículo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capítulo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('índice', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bibliografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resumen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consultar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subrayar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clasificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hemeroteca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extracto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('glosario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reseña', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ficha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catalogar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documentarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('meta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('logro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('éxito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fracaso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('constancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rendimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resultado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('superación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alcanzar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('progresar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perseverar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evolución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cumplimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balance', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aspiración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desempeño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('madurez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culminación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delegación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('administrativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acreditación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tasa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparecencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compulsar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('validación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fedatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diligencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsanación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('registrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gestoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('burocracia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('funcionariado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tramitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('folio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legajo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('providencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emplazamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventanilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('administración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extranjería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repatriación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expatriado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('migrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solicitante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empadronarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('naturalizarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arraigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apátrida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refugiarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temporal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permanente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reagrupación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biométrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prórroga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inadmisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deportación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nacionalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regularización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renovable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caducidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('padrón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frontera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consulado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gravamen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiscalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contribuyente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('liquidación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inspección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asegurado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siniestro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tributación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deducible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imponible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recaudación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hacienda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('copago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('franquicia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solvencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('morosidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recargo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deducción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseguradora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prima', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cobertura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuota', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saldo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrendador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrendatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subarriendo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usufructo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aval', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desperfecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inventario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rescisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vigencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrendamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cláusula', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habitabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moroso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desahucio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inmueble', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocupante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notarial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escriturar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usufructuario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canon', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('copropiedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('membrete', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acuse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enmienda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('borrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encabezado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diligenciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rubricar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protocolizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('registrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsanar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transcribir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('requerimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjunto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asunto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suplico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expongo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suscripción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rubricado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sellado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('foliación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imputado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('infracción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('querella', '', 'es_ru', CURRENT_TIMESTAMP),
  ('custodia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auxilio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('socorrer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('denuncia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testigo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('víctima', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amenaza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flagrancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hurto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estafa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atestado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arresto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('penal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('civil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('juicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peritaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clínica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especialista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diagnóstico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambulatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auscultar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diagnosticar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingresar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('triaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derivar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revisarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('examinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auscultación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cabecera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pediatra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('traumatólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dermatólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('analítica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfermería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cardiólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ginecólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('otorrino', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fractura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quemadura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hinchazón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inflamación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sangrado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esguince', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hematoma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calambre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jaqueca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cicatriz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entumecimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('punzada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empeorar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ampolla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moretón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('picor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rigidez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contractura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('luxación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rasguño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desmayo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalofrío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palpitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medicamento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dosis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antibiótico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pomada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reposo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prospecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pauta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aliviar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prevenir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aplicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dosificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinfectar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('analgésico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antiinflamatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprimido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cápsula', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vendaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hidratar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inhalador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('termómetro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraindicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apósito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mantenimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instalación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('garantía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pieza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arreglo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sustituir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ajustar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fontanero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electricista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('albañil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taladro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tornillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tuerca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grifo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persiana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cerradura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cañería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desagüe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interruptor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incendio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caída', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alarma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vigilancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prevención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extintor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evacuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortocircuito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escape', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intoxicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resbalón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proteger', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vigilar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escapar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asegurar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('botiquín', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barandilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antideslizante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mascarilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precaución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cerrojo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventilación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extinción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinfectante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('defecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('honesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('creativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sincero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impulsivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prudente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sociable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humilde', '', 'es_ru', CURRENT_TIMESTAMP),
  ('optimista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pesimista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('leal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maduro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testarudo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desconfiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambicioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reflexivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atrevido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introvertido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extrovertido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfeccionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rencoroso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ansiedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rabia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vergüenza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orgullo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alivio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entusiasmo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decepción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esperanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frustración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emoción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('angustia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irritación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ternura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nostalgia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('celos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('euforia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegrarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preocuparse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tranquilizarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agobio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('serenidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inquietud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desánimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('satisfacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cariño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vínculo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convivencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intimidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lealtad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconciliación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('complicidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cercanía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('afecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ruptura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acompañar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escuchar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parentesco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('noviazgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matrimonio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('separación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vecindad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cooperación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('afectivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprecio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distanciamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reproche', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culpa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ceder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alejarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consentir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tolerar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pactar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confrontación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distanciarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ofender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resentimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invasión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconciliarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vulnerar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interrumpir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quejarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reprochar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zanjar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desahogarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desautorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incomodar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apaciguar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pactación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estrés', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recompensa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relajación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('motivación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descanso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('energía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cambio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equilibrio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('organizarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concentrarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recuperarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('priorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('postergar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agotarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autocontrol', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perseverancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agotamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobrecarga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estímulo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('voluntad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autocuidado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relajarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desmotivación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resistencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regularidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procrastinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desconexión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobrellevar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respiración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sociedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('política', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gobierno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ley', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciudadano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('debate', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encuesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crisis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reforma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prensa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reportaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corresponsal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrevistar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('audiencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suceso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portavoz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manifestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('campaña', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crónica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rueda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guion', '', 'es_ru', CURRENT_TIMESTAMP),
  ('género', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estreno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('narrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protagonista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recomendar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reseñar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adaptar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rodar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temporada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('episodio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('largometraje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortometraje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comedia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('drama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suspense', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doblaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subtítulo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('banda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tráiler', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('montaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pintura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escultura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('público', '', 'es_ru', CURRENT_TIMESTAMP),
  ('espectáculo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instrumento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('melodía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ritmo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estrenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lienzo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orquesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recital', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acústico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compositor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bailarina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inauguración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mural', '', 'es_ru', CURRENT_TIMESTAMP),
  ('partitura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valoración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estilo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puntuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elogio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decepcionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sorprender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valorar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memorable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agradable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flojo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excelente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediocre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convincente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entretenido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emocionante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('logrado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irregular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previsible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('original', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impactante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decepcionante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recomendable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('voz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imagen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('textura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estética', '', 'es_ru', CURRENT_TIMESTAMP),
  ('criterio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convencer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apreciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subjetivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('objetivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coherente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atractivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expresivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sutil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exagerado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repetitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auténtico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('artificial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('profundidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('superficialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('armonía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preferible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disfrutable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ahorro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('préstamo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('riqueza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pobreza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ahorrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gastar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calcular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('endeudarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invertir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('financiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('liquidez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nómina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bruto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imprevisto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recorte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doméstica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrimonio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desembolso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('familiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('déficit', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doméstico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excedente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('económico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('producto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devolución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vendedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('etiqueta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('material', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reclamar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('duradero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('defectuoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embalaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reembolso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sustitución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprobante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acabado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incorrecta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descosido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rayado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manchado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('roto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caducado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comercial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gratuito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clave', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retirada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancaria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cargo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('validar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('banca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pin', '', 'es_ru', CURRENT_TIMESTAMP),
  ('débito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crédito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('domiciliación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reintegro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('beneficiario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bizum', '', 'es_ru', CURRENT_TIMESTAMP),
  ('domiciliado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secreta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('móvil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rebaja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demanda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('margen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('promoción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escaparate', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catálogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encarecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abaratar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ofertar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remate', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cupón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mayorista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('minorista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regateo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobreprecio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('puja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rebajado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('engaño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soporte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compensación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compensar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reemplazar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reembolsar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estafar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incumplimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vencida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('publicidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('engañosa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reclamaciones', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abierta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dañado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cobro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indebido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('erróneo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('injustificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deficiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('itinerario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transporte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excursión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escala', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alojarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recorrer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trayecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('folleto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previsto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ligero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guiada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pernoctar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reservarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anticipar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presupuestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empaquetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escapada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atasco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huelga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embarcar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retrasarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atascarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desviar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ferroviaria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('overbooking', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cancelarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desviarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demorado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('averiarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embotellamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colapso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retrasado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cancelado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transbordar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reprogramar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ferroviario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('infraestructura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comercio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('accesible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('céntrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peatonal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periférico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cercano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbanizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renovar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distrito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equipamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alumbrado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carril', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alcantarillado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acerado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carrilbici', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbanización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arbolado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('señalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contaminación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reciclaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residuo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sequía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sostenibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conservar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reciclar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contaminar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paisaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biodiversidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vertido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humareda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('erosión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reforestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biodegradable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atmosférico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nuboso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vivienda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edificio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mudanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reformar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convivir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derrama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vecinal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vecindario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('copropietario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rellano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humedades', '', 'es_ru', CURRENT_TIMESTAMP),
  ('domiciliario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('azotea', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fachada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('belleza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tranquilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encanto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descubrir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disfrutar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admirar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('panorámica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mirador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pintoresco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acogedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vibrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('turístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inolvidable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apacible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bullicioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descuidado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('animado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('silencioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monumental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dispositivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('configuración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('batería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cargador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ajuste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sistema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adaptador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cargarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enchufar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emparejar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bluetooth', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inalámbrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('táctil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('almacenamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brillo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volumen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auricular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cargable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calibrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inteligente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('funda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('router', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autenticación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('configurar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iniciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restablecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cifrado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autenticador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biometría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desbloqueo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('código', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doble', '', 'es_ru', CURRENT_TIMESTAMP),
  ('factor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antiphishing', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limitado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('token', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digital', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permisos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('identidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caducar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anonimato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suplantación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intruso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sospechoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desbloquear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desarrollo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('base', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repositorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entorno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diseñar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplegar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compilar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depurar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('backend', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frontend', '', 'es_ru', CURRENT_TIMESTAMP),
  ('script', '', 'es_ru', CURRENT_TIMESTAMP),
  ('framework', '', 'es_ru', CURRENT_TIMESTAMP),
  ('librería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('commit', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bug', '', 'es_ru', CURRENT_TIMESTAMP),
  ('técnica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despliegue', '', 'es_ru', CURRENT_TIMESTAMP),
  ('local', '', 'es_ru', CURRENT_TIMESTAMP),
  ('virtual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('continua', '', 'es_ru', CURRENT_TIMESTAMP),
  ('api', '', 'es_ru', CURRENT_TIMESTAMP),
  ('endpoint', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maqueta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prototipo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('componente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compilación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descarga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respaldo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exportación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('directorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprimir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descomprimir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('migrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restaurar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metadato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('binario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plantilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sincronización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remota', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automática', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('externo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restauración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dataset', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sincronizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exportable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('importable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('duplicado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incremental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instrucción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reinicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reportar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('error', '', 'es_ru', CURRENT_TIMESTAMP),
  ('captura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tutorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procedimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parche', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contactar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solucionador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('congelada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('forzado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ticket', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asistencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depuración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pantallazo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solucionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reiniciable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alerta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reintentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('directo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('difusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('privado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hilo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visualización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retransmitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grabación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retransmisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('creador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fijado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('métrica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('post', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multimedia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('streaming', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transmisor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suscriptor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('creadora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retransmisor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reaccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('difundir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('argumentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('premisa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('razonamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inducción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraargumento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('réplica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coherencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cohesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('postura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fundamento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perspectiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refutar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sostener', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demostrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conclusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lógica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('central', '', 'es_ru', CURRENT_TIMESTAMP),
  ('razonable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('debatible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sólido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consistente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inconsistente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refutación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demostración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inferir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deducir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inducir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fundamentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('articulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refutable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ilación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('silogismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseveración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('defendible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('argumentativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraejemplo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lógico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obligación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('necesidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prohibición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eventualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('factible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obligatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('opcional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admisible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ineludible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imprescindible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conveniente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contingente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exigible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permitido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prohibido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prever', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estimar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conceder', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descartar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obligatoriedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presunción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admisibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('factibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permisividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imperativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previsibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('riesgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asumible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('origen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cadena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proceso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('variable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correlación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desencadenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('influir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repercutir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incidir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('causar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propiciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('producir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multiplicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agravar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atenuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concatenación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('causal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nexo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indirecta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ramificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retroalimentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secundaria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dominó', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acumulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condicionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catalizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detonación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propagación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ramificarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intensificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desencadenamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambigüedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('complejidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paradoja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contradicción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuestionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suavizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('objetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problematizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('complejizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contextualizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restringir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atenuante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cautela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metodológica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terminológica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contralectura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alternativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sesgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simplificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reduccionismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobregeneralización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambivalencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concesivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problemático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discutible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('objecionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rebatible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discutibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cauteloso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condicionado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revisable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persuasión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intervención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interlocutor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deliberación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asamblea', '', 'es_ru', CURRENT_TIMESTAMP),
  ('foro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persuadir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intervenir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discrepar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transigir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciliar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interlocución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deliberar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deliberativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciliación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arbitraje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrapropuesta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palabra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('común', '', 'es_ru', CURRENT_TIMESTAMP),
  ('punto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muerto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consensuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disentir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('argumentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rebatir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polemizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pública', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciliador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regatear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deliberante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discrepante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociadora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraparte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neutralidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortesía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declaración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protocolario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('institucional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neutral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respetuoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manifestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declarar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impersonal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nominalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atenuador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encabezamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atentamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cordialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compareciente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suscrito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interesado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procedente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pertinente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conforme', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preceptivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reglado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acreditar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cumplimentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tramitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elevar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diligenciado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protocolizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('síntesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recapitulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('panorama', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hallazgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enseñanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moraleja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('global', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provisional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concluir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recapitular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proyectar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recapitulativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conclusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('argumental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corolario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('principal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derivada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('general', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraíble', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sintetizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compendio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('epílogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recapitulador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('finalizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desenlace', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discursivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condensación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clausura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eficiencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eficacia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('utilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alcance', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mejora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estándar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('optimización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excelencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventaja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desventaja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('optimizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('productivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eficiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprovechamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('satisfactorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('satisfactoriedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('idoneidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robustez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('funcionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consistencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('óptimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('real', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tangible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('optimizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rentable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sostenible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robusto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('funcional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operatividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprobable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('idóneo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tendencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transformación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crecimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declive', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retroceso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('innovación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adaptación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modernización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acelerar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ralentizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evolucionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transformar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consolidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mutación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('giro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renovación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expansión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contracción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estancamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deterioro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('progresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconfiguración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplazamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auge', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desaceleración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aceleración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fluctuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('variación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viraje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maduración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consolidación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gradualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desafío', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obstáculo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitigación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vulnerabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reducir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('afrontar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detectar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contingencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crítico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abordaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dificultad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('latente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grave', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resolución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitigable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neutralizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('minimizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contener', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preventiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escollo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contratiempo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fragilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emergente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsanable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reversible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irreversible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paliar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remediable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evitable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prevenible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desactivable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saneable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ética', '', 'es_ru', CURRENT_TIMESTAMP),
  ('justicia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('igualdad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('libertad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dignidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solidaridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('honestidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transparencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mérito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('virtud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('social', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cívico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rectitud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bien', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cívica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corresponsabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intelectual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colectiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legitimidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moralidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deontología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('altruismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('civismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imparcialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('honradez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humana', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distributiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('percepción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interpretación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actitud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mirada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sentido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('significado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('percibir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subjetividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('observador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('receptividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apreciación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subjetiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predisposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('receptivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interpretativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbólico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('connotación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emocional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cognitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('observación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apreciativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('externa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbólica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subjetivismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('observacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hermenéutica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interpretabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predisponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('percibido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('percibible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renuncia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jerarquía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sacrificio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('opción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seleccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equilibrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ponderar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escoger', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secundario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jerarquización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oportunidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mutua', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prelación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decisivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estratégica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prioritaria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ponderación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('priorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jerarquizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inestable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compensatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renunciable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prioritario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prescindible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preferente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsidiario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intercambiable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compensable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subóptimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('raíz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('núcleo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('motor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('freno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impulso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barrera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('horizonte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('luz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sombra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rumbo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tejido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palanca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ancla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brújula', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grieta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('umbral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suelo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('columna', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pilar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('engranaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bisagra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resorte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lastre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trampolín', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interna', '', 'es_ru', CURRENT_TIMESTAMP),
  ('timón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nudo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cauce', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corriente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pivote', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vertebral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telaraña', '', 'es_ru', CURRENT_TIMESTAMP),
  ('andamiaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('esqueleto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gestión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejecución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desviación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iteración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delegar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuenciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calendarizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('controlar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monitorear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alinear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replanificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dimensionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('priorizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estimativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entregabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reasignar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrocinador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gobernanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subproyecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trazabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intermedio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transversal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dimensionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuenciación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calendarización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ágil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retrospectiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refinamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estimable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comprometido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bloqueante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desviable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replanificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrocinio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reasignable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coordinación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejecutiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matricial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iterativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trazador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejecutor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejecutable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reasignación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('talento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mentor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconocimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('participación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dirigir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inspirar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empoderar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('motivar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mentoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('horizontal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('situacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multidisciplinar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pertenencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facilitador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cohesionador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delegante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('motivacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colaborativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('participativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corresponsable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mentorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tutelar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supervisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acompañamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empoderamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colectivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escucha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('activa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compartido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nuclear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ampliado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sinergia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alineamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cultural', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dinamizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cohesionante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('efectiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inclusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grupal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('visión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('misión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posicionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alineación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('táctico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('táctica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('directriz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estratégico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('competitiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('accionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuadro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mando', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('foco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuantificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corporativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('direccionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('competitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posicionador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuantificables', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proyección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operativizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instrumentalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplegable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('focalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diferenciación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compartida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prospectiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estratega', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parametrizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corporativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('direccionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('direccionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('organizacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuantificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tabulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operativas', '', 'es_ru', CURRENT_TIMESTAMP),
  ('venta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segmento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fidelización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('captación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posicionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prospección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embudo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prospector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posventa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cartera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('potencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segmentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fidelizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('captar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retener', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monetizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compraventa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preventa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comercialización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mercadeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comisionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clientela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nicho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('publicitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pujar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarifar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revendedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distribuidor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ofertante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demandante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monetización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comercializable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('promocionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comerciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comercializador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ofertador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('captador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derecho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confidencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('litigio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estipular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indemnizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rescindir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prorrogar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anexo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('penalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exclusividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jurisdicción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contractual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contractualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estipulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indemnización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('penalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rescindible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prorrogable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vinculante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vinculabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cesionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cedente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licenciatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licenciante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subcontrata', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subcontratista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjudicatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjudicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confidencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('territorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licenciamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adenda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adendado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cláusulado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rotación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incentivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salarial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('absentismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cantera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ascenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('laboral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('profesional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periódica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('voluntaria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desvinculación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('constructivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capacitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('individual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empleabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retribución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incentivable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relevo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('competencias', '', 'es_ru', CURRENT_TIMESTAMP),
  ('actitudinal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desvincular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recolocación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evaluabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empleable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('promocionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cualitativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuantitativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disputa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arbitrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confrontar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desescalar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desescalada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facilitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transacción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persistente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colaborativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pactada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negociabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciliable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transaccional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arbitral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contendiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfrentamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impasse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transigencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avenencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avenible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('litigioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('litigiosidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrapuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polarización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hostilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confrontativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conciliatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impugnación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impugnar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contencioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('videoconferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asincronía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('internacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flexible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distribuido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nómada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transfronterizo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multinacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intercultural', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remotamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desfase', '', 'es_ru', CURRENT_TIMESTAMP),
  ('itinerante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('híbrido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trabajo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('externalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('internacionalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teletrabajador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expatriación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslocalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telepresencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teleconferencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multinacionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plurilingüe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bicultural', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transnacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('externalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subcontratado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('movilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transfronteriza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arquitectura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('módulo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automatización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digitalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('microservicio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implementación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mantenibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desacoplar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encapsular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modularizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('versionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refactorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orquestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interoperable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acoplamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monolito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modularidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abstracción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arquitectónico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extensible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testeable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mantenible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desacoplado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modularizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refactorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encapsulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interfazado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('versionado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empaquetado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interoperabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extensibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integrabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arquitecturable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('componibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modularizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('portable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('integrable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('versionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testeabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conjunto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modelo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('algoritmo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predicción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clasificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procesar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('varianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('covarianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matriz', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distribución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agrupamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parámetro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estimador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muestreo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('etiquetado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobreajuste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predictor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predictivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vectorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incrustación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clasificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clusterización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dimensionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperparámetro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('optimizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convergencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('validable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrenable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inferencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probabilístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neuronal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supervisado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tokenización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estadística', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tokenizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vectorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probabilística', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bayesiano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inferidor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('predictibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vectorizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lingüístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ataque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filtración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auditoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intrusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anonimización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brecha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortafuegos', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciberseguridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encriptación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encriptar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descifrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atacante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explotable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explotación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcheado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parchear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('securizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('securización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trazable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auditable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auditabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seudonimización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anonimizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('credencialización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortafuego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antivirus', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antimalware', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intrusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('malicioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('securizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verificabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auditado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trazado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clúster', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monitorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automatizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balanceador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orquestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenedorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemetría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('observabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('latencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nodo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprovisionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respaldar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('virtualización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orquestador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprovisionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenedorizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('continuo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balanceo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redundancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tolerancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monitorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provisionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenerizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('virtualizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escalador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('orquestable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aprovisionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monitorizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redundante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balanceado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('virtualizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contenerizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automatizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplegador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despliegable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elasticidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concurrencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persistencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('experimento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('laboratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replicabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reproducibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('validez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protocolo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metodología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ensayo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('controlado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aleatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('falsabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('experimental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('longitudinal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuantitativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cualitativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empírico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metodológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reproducible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('replicable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('falsable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aleatorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('placebo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metaanálisis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('independiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('experimentalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operacionalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diseño', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrastación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muestreal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bibliográfica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('científica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('triangulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cohortal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exploratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('confirmatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explicativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iterar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('accesibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('navegación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flujo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prototipado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('navegabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('usable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intuitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conversacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fluidez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('activación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('heurística', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('microcopia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prototipable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('navegable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intuitividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('activable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retenible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convertible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descubribilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adoptabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prototipar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rediseñar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simplificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('experimentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('personalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('internacionalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segmentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('activar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convertir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instrumentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instrumentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('navegacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persuasivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('severidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degradación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reproducir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('investigar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degradado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degradar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anomalía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('observable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('remediación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcheo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regresivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('criticidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('severo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intermitente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monitoreo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incidentología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recuperable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fallar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degradable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diagnosticable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resolutivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitigador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correctivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preventivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estabilizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estabilización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reproceso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reprocesar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reintento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reabrir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reapertura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reincidencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('información', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inclusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exclusión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciudadanía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alfabetización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrónico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cibersociedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conectividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('algorítmico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tecnosociedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cibercultura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desigualdad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciberespacio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciberciudadanía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digitalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informatizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informatizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teleasistencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemedicina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teledemocracia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tecnopolítica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tecnopolítico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperconectado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperconectividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('institución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ministerio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parlamento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gabinete', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presidencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legislatura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejecutivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legislativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('judicial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('constitucional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('constitución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soberanía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descentralización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('centralización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atribución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mandato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('organismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('funcionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('institucionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gobernabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estatal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('republicano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('federal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('unitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('municipal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provincial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autonómico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ministerial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parlamentario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gubernamental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jurisdiccional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regulador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regulatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('centralizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interinstitucional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('potestad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('voto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('votante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urna', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sondeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electorado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escrutinio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('papeleta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circunscripción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abstención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abstencionismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propaganda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitin', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coalición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('oficialismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pluralismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ideología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conservador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('progresista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('liberal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('socialista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('centrista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('populista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electoral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elector', '', 'es_ru', CURRENT_TIMESTAMP),
  ('partidario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bipartidismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multipartidismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('primarias', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balotaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sufragio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('referendo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plebiscito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presidenciable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mayoritario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tribunal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('juez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiscal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sentencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apelación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acusación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('defensa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dictamen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escritura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procesal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mercantil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nulo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anulable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recurrible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apelable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condenatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('absolutorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cautelar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('probanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prescripción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embargo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('decomiso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('allanamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('homologación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ratificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jurisprudencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('casación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suplicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('auto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobreseimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imputación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imputar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('absolver', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condenar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demandado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('querellante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('querellado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('migración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emigración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inmigración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inmigrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emigrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asilo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refugio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refugiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('naturalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expulsión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diáspora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tránsito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embajada', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('migratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asilado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documentado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indocumentado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nacionalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retornado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplazado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reasentamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acogida', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estatuto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('permanencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empadronamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nacionalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repatriar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deportar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expulsar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regularizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fronterizo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estrato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('minoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mayoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('élite', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discriminación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segregación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marginalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('privilegio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precariedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('población', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('etnia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('raza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discapacidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diversidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interseccionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estigma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('segregado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marginal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vulnerable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('privilegiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('excluido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('minoritario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('identitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('demográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('generacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('étnico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('racial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redistribución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redistributivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inequidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vulnerabilizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marginar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crimen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delincuencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fraude', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corrupción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('homicidio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('violencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agresor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detenido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrulla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condena', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cárcel', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impunidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('persecución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incautación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pandilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extorsión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuestro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrabando', '', 'es_ru', CURRENT_TIMESTAMP),
  ('blanqueo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soborno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cohecho', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intimidación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delictivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('criminal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('punitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carcelario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('policial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agravante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('flagrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procesado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recluso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prófugo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('victimario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('victimización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diplomacia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tratado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cumbre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cancillería', '', 'es_ru', CURRENT_TIMESTAMP),
  ('embajador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bilateral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multilateral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('unilateral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geopolítica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geoestrategia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bloque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adhesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('membresía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tregua', '', 'es_ru', CURRENT_TIMESTAMP),
  ('armisticio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diplomático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interestatal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intergubernamental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supranacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humanitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pacificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pacificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enviado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delegatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canciller', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acreditado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plenipotenciario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extradición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraditar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cooperante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multilateralismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('unilateralismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bilateralismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regionalismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supranacionalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interdependiente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geopolítico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geoestratégico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limítrofe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marítimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cola', '', 'es_ru', CURRENT_TIMESTAMP),
  ('burocrático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsidio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiscalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teletrámite', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compulsado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pliego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concesionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('instanciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tramitador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tramitado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('archivístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('registral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catastro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catastral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empadronar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empadronado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('censal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('censo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('certificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compulsa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('requerir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notificante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resolutorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sancionador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concesional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('concesionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licitador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licitante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjudicador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjudicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adjudicable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('licitatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contratista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('macroeconomía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inflación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deflación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desempleo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empleo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consumo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('importación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('superávit', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciclo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coyuntura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monetario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cambiario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arancelario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('industrial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agregado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recesivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expansivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inflacionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deflacionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('improductivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estanflación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('endeudamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insolvencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iliquidez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('competitividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('macroeconómico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('microeconómico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('superavitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deficitario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presupuestario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contracíclico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procíclico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anticíclico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rentabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dividendo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fondo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('capital', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pasivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipoteca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pensión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('jubilación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aportación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diversificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diversificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refinanciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amortizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amortización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compuesto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volatilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agresivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especulativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ilíquido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solvente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insolvente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bursátil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('accionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bonista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inversor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ahorrador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apalancamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apalancar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinvertir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revalorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revalorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('propiedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tasación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notaría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('condominio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inmobiliario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotecario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arrendaticio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habitable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inhabitable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amueblado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reformado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tasado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escriturado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotecar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subarrendar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desalojar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desalojo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('okupación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocupación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avalista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plusvalía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derribar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edificable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbanizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proindiviso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotecable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotecado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subarrendador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subarrendatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desalojable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ocupacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('okupa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desahuciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tributo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contribución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tramo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mínimo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pagador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empleador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asalariado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declarable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tributable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recaudatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tributario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('previsional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devengo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devengar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('liquidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cotizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aportar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desgravar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desgravación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impositivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('imponibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retenedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retenido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deducibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cotizante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contributivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoliquidación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoliquidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declarativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('declarante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiscalizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fiscalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inspeccionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inspeccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crediticio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asegurador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asegurable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impagado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('refinanciación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('garante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('garantizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descubierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coaseguro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reaseguro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('financiero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancarizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancarización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancarizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bancarrota', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prestamista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prestatario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acreedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deudor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotecante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prendario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prenda', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avalado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseguramiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asegurabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siniestralidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siniestral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cubierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobregiro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amortizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amortizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reasegurador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negocio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fijo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recurrente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('básico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monetizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paquetizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subvencionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsidiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('financiarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preciar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarifario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tarifado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costeable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costeado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paquetización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empaquetamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paquetizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comisionable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recurrencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cruzado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subvención', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsidización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('financiador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('financiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facturador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cobrable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diferido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prorrateo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prorratear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descontar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('burbuja', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desplome', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pánico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contagio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devaluación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depreciación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escasez', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quiebra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moratoria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rescate', '', 'es_ru', CURRENT_TIMESTAMP),
  ('austeridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especulador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repunte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobrecalentamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfriamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depresivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volátil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bajista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alcista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sistémico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('energético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alimentario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('choque', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perturbación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('turbulencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diferencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inflacionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinflación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinflacionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperinflación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperinflacionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('devaluatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('depreciatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apreciatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estanflacionario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recesionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraccionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reestructuración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arte', '', 'es_ru', CURRENT_TIMESTAMP),
  ('forma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('danza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('poesía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dramaturgia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tragedia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('animación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('performance', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grabado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('collage', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retrato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bodegón', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ópera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ballet', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sinfonía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sonata', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('narrativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lírico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dramático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pictórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escultórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cinematográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teatral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coreográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('literario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('poético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('novelístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ensayístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('documentalista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vanguardista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('realista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbolista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impresionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expresionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encuadre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antagonista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diálogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('monólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subtexto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbolismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metáfora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ironía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verosimilitud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intertextualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fílmico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estilístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediometraje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filmografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filmar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rodaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('doblar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subtitulado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subtitular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('montajista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escenografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escenógrafo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrapicado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('encuadrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('omnisciente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('narratología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('narratológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diegético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extradiegético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periodismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entradilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('boletín', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exclusiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('primicia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desmentido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rumor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viralidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fotoperiodismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('infografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gráfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corresponsalía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mediático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periodístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editorialista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('opinativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('informativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('investigativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensacionalista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrastable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('redactor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reportero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cronista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrevistador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('entrevistado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('columnista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('articulista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editorializar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editorialización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('titularizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('titularización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maquetación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maquetar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desinformación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conmemoración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testimonio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tradición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('genealogía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('linaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('civilización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('época', '', 'es_ru', CURRENT_TIMESTAMP),
  ('período', '', 'es_ru', CURRENT_TIMESTAMP),
  ('siglo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('antigüedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('modernidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contemporaneidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('revolución', '', 'es_ru', CURRENT_TIMESTAMP),
  ('guerra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posguerra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exilio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dictadura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('democracia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colonización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descolonización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historiografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historiador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arqueología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arqueológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrimonial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conmemorativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('testimonial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ancestral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fundacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memorialismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('memorialista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historicidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('historicista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ritual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ceremonia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('símbolo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emblema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('folclore', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procesión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carnaval', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('herencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autóctono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ceremonial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tradicional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ritualizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mestizaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arraigado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pertenecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conmemorar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transmitir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preservar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ritualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ritualismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('celebrante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festejo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festejar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conmemorador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procesional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carnavalesco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('folclórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('folclorista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('composición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estructura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('originalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('influencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contextualización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('criticar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elogiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insinuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrastar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('valorativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irónico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambiguo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sugerente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exégesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relectura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comentador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reinterpretación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reinterpretar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polisemia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polisémico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambivalente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('connotativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('denotación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('denotativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problematización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problematizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraargumentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recepcional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canonización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('canonizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contextualizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contextualizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intertextual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbolización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simbolizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metafórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alegórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subtextual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('industria', '', 'es_ru', CURRENT_TIMESTAMP),
  ('productora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distribuidora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exhibición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mecenazgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curaduría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repertorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taquilla', '', 'es_ru', CURRENT_TIMESTAMP),
  ('producción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gestor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('productor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exhibidor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('promotor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('patrocinable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('audiovisual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escénico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('musical', '', 'es_ru', CURRENT_TIMESTAMP),
  ('museístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curatorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('industrialización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('industrializado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culturalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('creatividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editorialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('editabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('publicable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exhibible', '', 'es_ru', CURRENT_TIMESTAMP),
  ('programable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festivalero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('festivalesco', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('museal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('musealización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('musealizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escenificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escenificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taquillaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taquillero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('melancolía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apatía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoestima', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inseguridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desconfianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saturación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deseo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rechazo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desapego', '', 'es_ru', CURRENT_TIMESTAMP),
  ('duelo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('trauma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resiliencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autorregulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introspección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introspectivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ansioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('angustiado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eufórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('frustrado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resentido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culpable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avergonzado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aliviado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nostálgico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('melancólico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irritable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sereno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agotado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insegurizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desregular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desregulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('somatización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('somatizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disociación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consentimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empatía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asertividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invalidación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manipulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reciprocidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expectativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('negativa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indisponibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explícito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implícito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asertivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invasivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manipulador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invalidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acercarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('limitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delimitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delimitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consensuado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunicabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comunicativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escuchante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('empatizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disculpable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reparador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reprochable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invalidante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manipular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('chantaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('traición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abuso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dominio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sumisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agresividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evasión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rencor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intimidar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amenazar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desconfiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sospechar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('traicionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culpabilizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('victimizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coerción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coercitivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('controlador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dominante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sumiso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hostil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('defensivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coaccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coactivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intimidatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('amenazante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manipulativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslegitimar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslegitimación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desacreditar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desacreditación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('victimismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('victimista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('culpabilización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crianza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('maternidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paternidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hogar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('divorcio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adopción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('padrastro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('madrastra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hijastro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suegro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('yerno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nuera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuñado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coparentalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('coparental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuidador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuidadora', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conyugal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fraternal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solidario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nupcial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conyugalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parentela', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filiación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('filial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consanguinidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consanguíneo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corresidencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('convivencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temperamento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conducta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comportamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('procrastinación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perfeccionismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impulsividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extraversión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('apertura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neuroticismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rasgo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoimagen', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoconcepto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autocrítica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoexigencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoobservación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('repetición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('automatismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compulsión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compulsivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rígido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('constante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('perseverante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disciplinado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habituarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deshabituarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adaptarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conductual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('temperamental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caracterológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extravertido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('neurótico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoeficacia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('autoconocimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desgaste', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insomnio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irritabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('meditación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bienestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('malestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fatiga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperactividad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('crónico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agudo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quemarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regularse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('meditar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobrecargado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exhausto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fatigado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desbordado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estresor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estresante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estresarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobrecargar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extenuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extenuado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('extenuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fatigar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hiperexigencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relajante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dosificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('balancear', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pausado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pausar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insomne', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irritar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recuperativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('restaurativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urgencias', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfermera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacunación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cribado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('epidemiología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('epidemiológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asistencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('clínico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanitarista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mutualidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambulatorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacunatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vacunador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inmunización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cribador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cribaje', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asistencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derivable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalizable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urgenciólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('facultativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanitarización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sanitarizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('derivador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambulatorial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambulatorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hospitalizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingresable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ingresado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('admisionista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('exploración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('radiografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resonancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tomografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biopsia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('endoscopia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrocardiograma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('signo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terapia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cirugía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anestesia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rehabilitación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fisioterapia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tratar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('medicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('operar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rehabilitar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('analítico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('radiológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('quirúrgico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terapéutico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anamnesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palpación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sintomatología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sintomático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('asintomático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('radiografiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecografiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biopsiar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('endoscópico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tomográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resonador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrocardiográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sedentarismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obesidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipertensión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diabetes', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colesterol', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tabaquismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alcoholismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adherencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recaída', '', 'es_ru', CURRENT_TIMESTAMP),
  ('complicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('brote', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comorbilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abandonar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mantener', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estabilizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('controlable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estabilizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('degenerativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metabólico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cardiovascular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respiratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cronicidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cronificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cronificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sedentario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipertenso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diabético', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipercolesterolemia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dislipidemia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tabaquista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nutrición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alimentación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proteína', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carbohidrato', '', 'es_ru', CURRENT_TIMESTAMP),
  ('grasa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fibra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vitamina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mineral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hidratación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fuerza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deportista', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caloría', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metabolismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digestión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('saciedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ayuno', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suplemento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suplementación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('muscular', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aeróbico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('anaeróbico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nutritivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calórico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digestivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hidratado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deshidratado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estirar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fortalecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hidratarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alimentarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teleconsulta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telesalud', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biosensor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prótesis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robótica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('radiología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('videoconsulta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('genómica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('genómico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bioinformática', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bioinformático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nanotecnología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nanomedicina', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conectado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implantable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protésico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robotizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemédico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('teleasistencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telediagnóstico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telediagnosticar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemonitorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemonitorizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('telemática', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robotización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robotizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('robótico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protetización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protetizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implantología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implantólogo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biosensórica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biosensorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('advertencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobredosis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intolerancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('toxicidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('toxicología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abstinencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('somnolencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('erupción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('farmacovigilancia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indicación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suspensión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aumentar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adverso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraindicado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alérgico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tóxico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inflamatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('leve', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contraindicar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precaucional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('posológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('interaccionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipersensibilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensibilización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sensibilizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intolerante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('nauseoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('somnoliento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irritativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hemorrágico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hemorragia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discontinuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retirar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pautar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terapéutica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('climático', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calentamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inundación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desertificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecosistema', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hábitat', '', 'es_ru', CURRENT_TIMESTAMP),
  ('especie', '', 'es_ru', CURRENT_TIMESTAMP),
  ('carbono', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deforestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acidificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambiental', '', 'es_ru', CURRENT_TIMESTAMP),
  ('regeneración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conservación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitigar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descontaminar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('descarbonizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('renaturalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fósil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hídrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('climatología', '', 'es_ru', CURRENT_TIMESTAMP),
  ('climatológico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('calentarse', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desertificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desertificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('erosionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('erosivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('biodiverso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ecosistémico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('invasor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contaminante', '', 'es_ru', CURRENT_TIMESTAMP),
  ('emisivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deforestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reforestar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acidificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbanismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciudad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zonificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('densidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peatonalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('altura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('edificabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gentrificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('periferia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suburbio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metrópoli', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conurbación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ordenamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mixto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('compacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('disperso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caminable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('urbanístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciclable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dotacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('densificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recalificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reurbanizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('zonificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('densificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('densificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('peatonalizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reurbanización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reurbanizador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ordenanza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planeamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('planeador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcelación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('parcelar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dotación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ciclovía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subestación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transmisión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eólico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hidroeléctrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recarga', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intermodalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('logística', '', 'es_ru', CURRENT_TIMESTAMP),
  ('túnel', '', 'es_ru', CURRENT_TIMESTAMP),
  ('viaducto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('corredor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ferrocarril', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eléctrico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrificado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intermodal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('logístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multimodal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abastecimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tranviario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metropolitano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electrificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electromovilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('electromóvil', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recargable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('baterización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fotovoltaico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solarizado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('multimodalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tunelización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soterramiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('soterrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abastecedor', '', 'es_ru', CURRENT_TIMESTAMP),
  ('región', '', 'es_ru', CURRENT_TIMESTAMP),
  ('territorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llanura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('meseta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('litoral', '', 'es_ru', CURRENT_TIMESTAMP),
  ('península', '', 'es_ru', CURRENT_TIMESTAMP),
  ('archipiélago', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuenca', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desierto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('selva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('humedal', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estepa', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tundra', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volcán', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cordillera', '', 'es_ru', CURRENT_TIMESTAMP),
  ('altiplano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bahía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('golfo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delta', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cartografía', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relieve', '', 'es_ru', CURRENT_TIMESTAMP),
  ('bioma', '', 'es_ru', CURRENT_TIMESTAMP),
  ('geográfico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paisajístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('costero', '', 'es_ru', CURRENT_TIMESTAMP),
  ('montañoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('fluvial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lacustre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desértico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('iluminación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proximidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('barrial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ruidoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contaminado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventilado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sombreado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('próximo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('equipado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deteriorado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rehabilitable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('habitacional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('residencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('vecinalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distrital', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ambientalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('salubridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insalubre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acústica', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sonoro', '', 'es_ru', CURRENT_TIMESTAMP),
  ('lumínico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ventilatorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arboladura', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arborización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('arborizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('caminabilidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deteriorar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('gentrificar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insalubridad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insonorización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catástrofe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desastre', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terremoto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huracán', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslizamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('avalancha', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tsunami', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alimento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconstrucción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('racionamiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resistente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('resiliente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agotable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('catastrófico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('desastroso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evacuar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evacuado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inundable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incendiario', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sísmico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('huracanado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tormentoso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deslizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('volcánico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eruptivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('colapsar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('escaso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reconstruir', '', 'es_ru', CURRENT_TIMESTAMP),
  ('suministrar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('racionar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('abastecer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('agotar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adaptativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reservorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('igualmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('análogamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simultáneamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('seguidamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ulteriormente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('consiguientemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sucintamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('detalladamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('específicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('particularmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metatexto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metatextual', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introducción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('epígrafe', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inciso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ejemplificación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reformulación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('expositivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('secuenciador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('introductorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aclaratorio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delimitador', '', 'es_ru', CURRENT_TIMESTAMP),
  ('transicional', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digresivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('digresión', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enumeración', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enumerativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tematización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tematizar', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contrariamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('inversamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('recíprocamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alternativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comparativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('correlativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('proporcionalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respectivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('adicionalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('complementariamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('subsidiariamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('preliminarmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provisionalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tentativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presumiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('plausiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('verosímilmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aparentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indudablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incuestionablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discutiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cuestionablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('paradójicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('significativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sustancialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('marginalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tangencialmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incidentalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('metodológicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conceptualmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('terminológicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('semánticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('epistemológicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hermenéuticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discursivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('pragmáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('normativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('críticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ontológicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('axiológicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evidentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('obviamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('naturalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indiscutiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('acaso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('supuestamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presuntamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('simplemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('meramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('solamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('únicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('precisamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('justamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('relativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('considerablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('profundamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('radicalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('moderadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cautelosamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('francamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('honestamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ostensiblemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('palmariamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('manifiestamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('notoriamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indudable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('hipotéticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conjeturalmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aproximativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobradamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('matizadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('implícitamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('explícitamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deliberadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('intencionadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('retóricamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('estratégicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('enfáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('taxativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('categóricamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rotundamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tajantemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prudentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('irónicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('llamativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('curiosamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sorprendentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reveladoramente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sintomáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elocuentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sugestivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('problemáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('polémicamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('controvertidamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('razonablemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('justificadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('legítimamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sesgadamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tendenciosamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('provocativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseverativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dubitativamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incisivamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contundentemente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impugnable', '', 'es_ru', CURRENT_TIMESTAMP),
  ('aseverativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('dubitativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('conjetural', '', 'es_ru', CURRENT_TIMESTAMP),
  ('incisivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('contundente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tacto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('delicadeza', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discreción', '', 'es_ru', CURRENT_TIMESTAMP),
  ('prudencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cordialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insinuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sugerencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eufemismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atenuación', '', 'es_ru', CURRENT_TIMESTAMP),
  ('rodeo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elipsis', '', 'es_ru', CURRENT_TIMESTAMP),
  ('silencio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobreentendido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presuposición', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insinuado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tácito', '', 'es_ru', CURRENT_TIMESTAMP),
  ('velado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deferente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discreto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('indirecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('atenuado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mitigado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elíptico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deferencial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortés', '', 'es_ru', CURRENT_TIMESTAMP),
  ('respetuosidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reticencia', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circunloquio', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evasiva', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sugestivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('despersonalización', '', 'es_ru', CURRENT_TIMESTAMP),
  ('impersonalidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('deferencialidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('protocolariedad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formulismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('formalismo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('cortesano', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ceremonioso', '', 'es_ru', CURRENT_TIMESTAMP),
  ('ceremoniosidad', '', 'es_ru', CURRENT_TIMESTAMP),
  ('distanciado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('reticente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circunloquial', '', 'es_ru', CURRENT_TIMESTAMP),
  ('evasivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('insinuativo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sobreentender', '', 'es_ru', CURRENT_TIMESTAMP),
  ('presuponer', '', 'es_ru', CURRENT_TIMESTAMP),
  ('tácitamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('veladamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('sutilmente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('diplomáticamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('discretamente', '', 'es_ru', CURRENT_TIMESTAMP),
  ('eufemístico', '', 'es_ru', CURRENT_TIMESTAMP),
  ('elusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('alusivo', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circunspecto', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comedido', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mesurado', '', 'es_ru', CURRENT_TIMESTAMP),
  ('circunspección', '', 'es_ru', CURRENT_TIMESTAMP),
  ('comedimiento', '', 'es_ru', CURRENT_TIMESTAMP),
  ('mesura', '', 'es_ru', CURRENT_TIMESTAMP)
ON CONFLICT(word) DO UPDATE SET
  course_code = 'es_ru',
  updated_at = CURRENT_TIMESTAMP;

-- A0 / Служебные слова и грамматический минимум / Личные местоимения и обращения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Личные местоимения и обращения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Личные местоимения и обращения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'yo'),
    (1, 'tú'),
    (2, 'usted'),
    (3, 'él'),
    (4, 'ella'),
    (5, 'nosotros'),
    (6, 'vosotros'),
    (7, 'ustedes')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Служебные слова и грамматический минимум / Указательные и притяжательные основы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Указательные и притяжательные основы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Указательные и притяжательные основы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'este'),
    (1, 'ese'),
    (2, 'aquel'),
    (3, 'mi'),
    (4, 'tu'),
    (5, 'su')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Служебные слова и грамматический минимум / Вопросительные слова
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Вопросительные слова'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Вопросительные слова'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'qué'),
    (1, 'quién'),
    (2, 'cuál'),
    (3, 'cuánto'),
    (4, 'cómo'),
    (5, 'cuándo'),
    (6, 'dónde'),
    (7, 'adónde')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Служебные слова и грамматический минимум / Самые частые предлоги и союзы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Самые частые предлоги и союзы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Самые частые предлоги и союзы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'por'),
    (1, 'de'),
    (2, 'en'),
    (3, 'con'),
    (4, 'sin'),
    (5, 'para'),
    (6, 'y'),
    (7, 'o')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Служебные слова и грамматический минимум / Частицы, согласие, отрицание, степень
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Частицы, согласие, отрицание, степень'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Служебные слова и грамматический минимум'
    AND ws.title = 'Частицы, согласие, отрицание, степень'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sí'),
    (1, 'no'),
    (2, 'muy'),
    (3, 'también'),
    (4, 'tampoco')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Люди и контакт / Приветствия, прощания, вежливость
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Приветствия, прощания, вежливость'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Приветствия, прощания, вежливость'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'hola'),
    (1, 'adiós'),
    (2, 'gracias'),
    (3, 'perdón'),
    (4, 'disculpa'),
    (5, 'bienvenido'),
    (6, 'saludo'),
    (7, 'chao')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Люди и контакт / Семья и близкие люди
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Семья и близкие люди'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Семья и близкие люди'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'familia'),
    (1, 'madre'),
    (2, 'padre'),
    (3, 'hijo'),
    (4, 'hija'),
    (5, 'hermano'),
    (6, 'hermana'),
    (7, 'abuelo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Люди и контакт / Страны, языки, национальности: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Страны, языки, национальности: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Страны, языки, национальности: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'país'),
    (1, 'idioma'),
    (2, 'nacionalidad'),
    (3, 'español'),
    (4, 'extranjero')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Люди и контакт / Простые роли: человек, друг, мужчина, женщина
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Простые роли: человек, друг, мужчина, женщина'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Люди и контакт'
    AND ws.title = 'Простые роли: человек, друг, мужчина, женщина'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'persona'),
    (1, 'amigo'),
    (2, 'hombre'),
    (3, 'mujer')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Базовые действия / Быть, находиться, иметь, делать
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Быть, находиться, иметь, делать'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Быть, находиться, иметь, делать'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ser'),
    (1, 'estar'),
    (2, 'tener'),
    (3, 'hacer'),
    (4, 'haber'),
    (5, 'existir'),
    (6, 'quedar'),
    (7, 'llamarse')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Базовые действия / Движение и повседневные действия
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Движение и повседневные действия'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Движение и повседневные действия'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ir'),
    (1, 'venir'),
    (2, 'caminar'),
    (3, 'entrar'),
    (4, 'salir'),
    (5, 'subir'),
    (6, 'bajar'),
    (7, 'abrir'),
    (8, 'cerrar'),
    (9, 'comer')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Базовые действия / Хотеть, мочь, знать, понимать, нужно
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Хотеть, мочь, знать, понимать, нужно'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Базовые действия'
    AND ws.title = 'Хотеть, мочь, знать, понимать, нужно'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'querer'),
    (1, 'poder'),
    (2, 'saber'),
    (3, 'entender'),
    (4, 'necesitar'),
    (5, 'deber'),
    (6, 'comprender')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Предметы вокруг / Дом и комната: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Дом и комната: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Дом и комната: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'casa'),
    (1, 'habitación'),
    (2, 'puerta'),
    (3, 'ventana'),
    (4, 'mesa'),
    (5, 'silla'),
    (6, 'cama'),
    (7, 'baño')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Предметы вокруг / Личные вещи и документы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Личные вещи и документы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Личные вещи и документы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'bolso'),
    (1, 'mochila'),
    (2, 'llave'),
    (3, 'dinero'),
    (4, 'pasaporte'),
    (5, 'documento'),
    (6, 'tarjeta')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Предметы вокруг / Телефон, учёба, интернет: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Телефон, учёба, интернет: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Предметы вокруг'
    AND ws.title = 'Телефон, учёба, интернет: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'teléfono'),
    (1, 'escuela'),
    (2, 'clase'),
    (3, 'internet'),
    (4, 'mensaje')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Еда, город, покупки / Еда и напитки: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Еда и напитки: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Еда и напитки: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'agua'),
    (1, 'pan'),
    (2, 'leche'),
    (3, 'café'),
    (4, 'té'),
    (5, 'carne'),
    (6, 'fruta'),
    (7, 'verdura')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Еда, город, покупки / Места в городе
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Места в городе'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Места в городе'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'calle'),
    (1, 'plaza'),
    (2, 'tienda'),
    (3, 'banco'),
    (4, 'hotel'),
    (5, 'restaurante')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Еда, город, покупки / Транспорт, покупка, оплата: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Транспорт, покупка, оплата: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Еда, город, покупки'
    AND ws.title = 'Транспорт, покупка, оплата: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'autobús'),
    (1, 'tren'),
    (2, 'taxi'),
    (3, 'billete'),
    (4, 'comprar'),
    (5, 'pagar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Числа, время, свойства / Числа: базовый набор
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Числа: базовый набор'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Числа: базовый набор'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cero'),
    (1, 'uno'),
    (2, 'dos'),
    (3, 'tres'),
    (4, 'cuatro'),
    (5, 'cinco'),
    (6, 'seis'),
    (7, 'siete'),
    (8, 'ocho'),
    (9, 'nueve'),
    (10, 'diez'),
    (11, 'cien')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Числа, время, свойства / Дни, части дня, сейчас/потом
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Дни, части дня, сейчас/потом'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Дни, части дня, сейчас/потом'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'día'),
    (1, 'mañana'),
    (2, 'tarde'),
    (3, 'noche'),
    (4, 'ahora')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A0 / Числа, время, свойства / Цвета, размер, базовые признаки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Цвета, размер, базовые признаки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A0'
    AND cat.name = 'Числа, время, свойства'
    AND ws.title = 'Цвета, размер, базовые признаки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'blanco'),
    (1, 'negro'),
    (2, 'rojo'),
    (3, 'azul'),
    (4, 'verde'),
    (5, 'grande'),
    (6, 'pequeño'),
    (7, 'nuevo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Служебные слова A1 / Местоимения дополнения: основы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Местоимения дополнения: основы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Местоимения дополнения: основы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'me'),
    (1, 'te'),
    (2, 'lo'),
    (3, 'la'),
    (4, 'le'),
    (5, 'nos'),
    (6, 'os'),
    (7, 'los'),
    (8, 'las'),
    (9, 'les'),
    (10, 'mí'),
    (11, 'ti')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Служебные слова A1 / Притяжательные, указательные, неопределённые
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Притяжательные, указательные, неопределённые'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Притяжательные, указательные, неопределённые'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'mío'),
    (1, 'tuyo'),
    (2, 'suyo'),
    (3, 'nuestro'),
    (4, 'vuestro'),
    (5, 'esto'),
    (6, 'eso'),
    (7, 'algo'),
    (8, 'alguien'),
    (9, 'nadie'),
    (10, 'todo'),
    (11, 'otro')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Служебные слова A1 / Предлоги места, направления, времени
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Предлоги места, направления, времени'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Предлоги места, направления, времени'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'hacia'),
    (1, 'desde'),
    (2, 'hasta'),
    (3, 'sobre'),
    (4, 'entre'),
    (5, 'contra'),
    (6, 'durante'),
    (7, 'antes'),
    (8, 'después'),
    (9, 'delante'),
    (10, 'detrás'),
    (11, 'cerca')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Служебные слова A1 / Союзы и связки простых фраз
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Союзы и связки простых фраз'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Союзы и связки простых фраз'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'pero'),
    (1, 'porque'),
    (2, 'cuando'),
    (3, 'si'),
    (4, 'aunque'),
    (5, 'mientras'),
    (6, 'entonces'),
    (7, 'pues'),
    (8, 'además')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Служебные слова A1 / Наречия частотности, степени, порядка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Наречия частотности, степени, порядка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Служебные слова A1'
    AND ws.title = 'Наречия частотности, степени, порядка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'siempre'),
    (1, 'nunca'),
    (2, 'aún'),
    (3, 'ya'),
    (4, 'pronto'),
    (5, 'luego'),
    (6, 'primero'),
    (7, 'solo'),
    (8, 'bastante'),
    (9, 'demasiado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Повседневные действия и рутина / Утро, вечер, бытовой день
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Утро, вечер, бытовой день'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Утро, вечер, бытовой день'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'despertar'),
    (1, 'levantarse'),
    (2, 'ducharse'),
    (3, 'lavarse'),
    (4, 'vestirse'),
    (5, 'desayunar'),
    (6, 'almorzar'),
    (7, 'cenar'),
    (8, 'dormir'),
    (9, 'descansar'),
    (10, 'trabajar'),
    (11, 'estudiar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Повседневные действия и рутина / Готовить, есть, покупать
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Готовить, есть, покупать'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Готовить, есть, покупать'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cocinar'),
    (1, 'beber'),
    (2, 'probar'),
    (3, 'pedir'),
    (4, 'servir'),
    (5, 'vender'),
    (6, 'elegir'),
    (7, 'costar'),
    (8, 'precio'),
    (9, 'mercado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Повседневные действия и рутина / Перемещаться, ехать, приходить
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Перемещаться, ехать, приходить'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Перемещаться, ехать, приходить'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'viajar'),
    (1, 'conducir'),
    (2, 'llegar'),
    (3, 'volver'),
    (4, 'pasar'),
    (5, 'cruzar'),
    (6, 'esperar'),
    (7, 'parar'),
    (8, 'correr'),
    (9, 'seguir')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Повседневные действия и рутина / Просить, давать, помогать, брать
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Просить, давать, помогать, брать'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Просить, давать, помогать, брать'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'dar'),
    (1, 'tomar'),
    (2, 'recibir'),
    (3, 'ayudar'),
    (4, 'buscar'),
    (5, 'encontrar'),
    (6, 'usar'),
    (7, 'llevar'),
    (8, 'dejar'),
    (9, 'traer')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Повседневные действия и рутина / Чувствовать, думать, помнить: основы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Чувствовать, думать, помнить: основы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Повседневные действия и рутина'
    AND ws.title = 'Чувствовать, думать, помнить: основы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sentir'),
    (1, 'pensar'),
    (2, 'creer'),
    (3, 'recordar'),
    (4, 'olvidar'),
    (5, 'gustar'),
    (6, 'preferir'),
    (7, 'parecer')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Дом, семья, личная информация / Анкета, контакты, личные данные
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Анкета, контакты, личные данные'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Анкета, контакты, личные данные'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'nombre'),
    (1, 'apellido'),
    (2, 'edad'),
    (3, 'dirección'),
    (4, 'correo'),
    (5, 'número'),
    (6, 'firma'),
    (7, 'formulario'),
    (8, 'contacto'),
    (9, 'dato')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Дом, семья, личная информация / Семья, возраст, семейный статус
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Семья, возраст, семейный статус'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Семья, возраст, семейный статус'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'esposo'),
    (1, 'esposa'),
    (2, 'pareja'),
    (3, 'nieto'),
    (4, 'nieta'),
    (5, 'tío'),
    (6, 'tía'),
    (7, 'primo'),
    (8, 'joven'),
    (9, 'casado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Дом, семья, личная информация / Жильё, комнаты, мебель
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Жильё, комнаты, мебель'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Жильё, комнаты, мебель'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'piso'),
    (1, 'apartamento'),
    (2, 'cocina'),
    (3, 'salón'),
    (4, 'dormitorio'),
    (5, 'pasillo'),
    (6, 'sofá'),
    (7, 'armario'),
    (8, 'estantería'),
    (9, 'ducha'),
    (10, 'pared'),
    (11, 'techo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Дом, семья, личная информация / Одежда и внешность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Одежда и внешность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Одежда и внешность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ropa'),
    (1, 'camisa'),
    (2, 'pantalón'),
    (3, 'zapato'),
    (4, 'abrigo'),
    (5, 'vestido'),
    (6, 'alto'),
    (7, 'bajo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Дом, семья, личная информация / Базовые документы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Базовые документы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Дом, семья, личная информация'
    AND ws.title = 'Базовые документы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'carné'),
    (1, 'permiso'),
    (2, 'visado'),
    (3, 'certificado'),
    (4, 'copia')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Еда, покупки, услуги / Продукты и напитки A1
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Продукты и напитки A1'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Продукты и напитки A1'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'arroz'),
    (1, 'pasta'),
    (2, 'queso'),
    (3, 'huevo'),
    (4, 'pollo'),
    (5, 'pescado'),
    (6, 'sopa'),
    (7, 'ensalada'),
    (8, 'manzana'),
    (9, 'plátano'),
    (10, 'naranja'),
    (11, 'azúcar'),
    (12, 'sal'),
    (13, 'aceite'),
    (14, 'zumo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Еда, покупки, услуги / Кафе, ресторан, меню
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Кафе, ресторан, меню'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Кафе, ресторан, меню'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'menú'),
    (1, 'plato'),
    (2, 'camarero'),
    (3, 'cuenta'),
    (4, 'propina'),
    (5, 'reserva'),
    (6, 'postre'),
    (7, 'bebida'),
    (8, 'taza'),
    (9, 'vaso')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Еда, покупки, услуги / Магазин, цена, оплата
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Магазин, цена, оплата'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Магазин, цена, оплата'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cliente'),
    (1, 'cajero'),
    (2, 'caja'),
    (3, 'descuento'),
    (4, 'oferta'),
    (5, 'recibo'),
    (6, 'efectivo'),
    (7, 'moneda'),
    (8, 'caro'),
    (9, 'barato'),
    (10, 'probarse'),
    (11, 'devolver')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Еда, покупки, услуги / Услуги: аптека, банк, почта
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Услуги: аптека, банк, почта'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Услуги: аптека, банк, почта'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'farmacia'),
    (1, 'medicina'),
    (2, 'envío'),
    (3, 'paquete'),
    (4, 'sello'),
    (5, 'transferencia'),
    (6, 'sucursal'),
    (7, 'servicio')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Еда, покупки, услуги / Упаковка, количество, единицы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Упаковка, количество, единицы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Еда, покупки, услуги'
    AND ws.title = 'Упаковка, количество, единицы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'botella'),
    (1, 'envase'),
    (2, 'bolsa'),
    (3, 'cartón'),
    (4, 'lata'),
    (5, 'kilo'),
    (6, 'litro'),
    (7, 'gramo'),
    (8, 'trozo'),
    (9, 'docena')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Город, транспорт, путешествие / Места в городе A1
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Места в городе A1'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Места в городе A1'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'parque'),
    (1, 'museo'),
    (2, 'cine'),
    (3, 'teatro'),
    (4, 'hospital'),
    (5, 'oficina'),
    (6, 'estación'),
    (7, 'aeropuerto'),
    (8, 'puerto'),
    (9, 'biblioteca'),
    (10, 'iglesia'),
    (11, 'comisaría')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Город, транспорт, путешествие / Общественный транспорт
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Общественный транспорт'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Общественный транспорт'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'metro'),
    (1, 'tranvía'),
    (2, 'parada'),
    (3, 'andén'),
    (4, 'conductor'),
    (5, 'pasajero'),
    (6, 'horario'),
    (7, 'ruta'),
    (8, 'abono'),
    (9, 'asiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Город, транспорт, путешествие / Направления и маршрут
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Направления и маршрут'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Направления и маршрут'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'izquierda'),
    (1, 'derecha'),
    (2, 'recto'),
    (3, 'esquina'),
    (4, 'mapa'),
    (5, 'camino'),
    (6, 'orientación'),
    (7, 'norte'),
    (8, 'sur'),
    (9, 'centro')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Город, транспорт, путешествие / Гостиница и аренда: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Гостиница и аренда: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Гостиница и аренда: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'alojamiento'),
    (1, 'recepción'),
    (2, 'huésped'),
    (3, 'alquilar'),
    (4, 'alquiler'),
    (5, 'llavero')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Город, транспорт, путешествие / Багаж, билеты, поездка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Багаж, билеты, поездка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Город, транспорт, путешествие'
    AND ws.title = 'Багаж, билеты, поездка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'maleta'),
    (1, 'equipaje'),
    (2, 'viaje'),
    (3, 'ida'),
    (4, 'vuelta'),
    (5, 'pasaje'),
    (6, 'destino')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Время, погода, календарь / Дни, месяцы, сезоны
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Дни, месяцы, сезоны'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Дни, месяцы, сезоны'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'lunes'),
    (1, 'martes'),
    (2, 'miércoles'),
    (3, 'jueves'),
    (4, 'viernes'),
    (5, 'sábado'),
    (6, 'domingo'),
    (7, 'enero'),
    (8, 'febrero'),
    (9, 'marzo'),
    (10, 'verano'),
    (11, 'invierno')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Время, погода, календарь / Часы и расписание
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Часы и расписание'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Часы и расписание'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'hora'),
    (1, 'minuto'),
    (2, 'segundo'),
    (3, 'reloj'),
    (4, 'calendario'),
    (5, 'fecha'),
    (6, 'cita'),
    (7, 'reunión'),
    (8, 'turno'),
    (9, 'puntual')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Время, погода, календарь / Частотность и длительность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Частотность и длительность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Частотность и длительность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'diario'),
    (1, 'semanal'),
    (2, 'mensual'),
    (3, 'anual'),
    (4, 'frecuente'),
    (5, 'raro'),
    (6, 'duración'),
    (7, 'vez')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Время, погода, календарь / Погода A1
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Погода A1'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Погода A1'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'tiempo'),
    (1, 'sol'),
    (2, 'lluvia'),
    (3, 'nieve'),
    (4, 'viento'),
    (5, 'nube'),
    (6, 'calor'),
    (7, 'frío')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Время, погода, календарь / Числа 20-100 и порядок
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Числа 20-100 и порядок'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Время, погода, календарь'
    AND ws.title = 'Числа 20-100 и порядок'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'veinte'),
    (1, 'treinta'),
    (2, 'cuarenta'),
    (3, 'cincuenta'),
    (4, 'sesenta'),
    (5, 'setenta'),
    (6, 'último')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Описание и состояния / Прилагательные для людей
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Прилагательные для людей'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Прилагательные для людей'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'simpático'),
    (1, 'antipático'),
    (2, 'amable'),
    (3, 'serio'),
    (4, 'tranquilo'),
    (5, 'nervioso'),
    (6, 'feliz'),
    (7, 'triste'),
    (8, 'cansado'),
    (9, 'ocupado'),
    (10, 'libre'),
    (11, 'listo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Описание и состояния / Прилагательные для вещей
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Прилагательные для вещей'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Прилагательные для вещей'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'viejo'),
    (1, 'moderno'),
    (2, 'bonito'),
    (3, 'feo'),
    (4, 'limpio'),
    (5, 'sucio'),
    (6, 'fácil'),
    (7, 'difícil'),
    (8, 'rápido'),
    (9, 'lento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Описание и состояния / Эмоции и самочувствие
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Эмоции и самочувствие'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Эмоции и самочувствие'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'alegría'),
    (1, 'tristeza'),
    (2, 'miedo'),
    (3, 'enfado'),
    (4, 'dolor'),
    (5, 'hambre'),
    (6, 'sed'),
    (7, 'sueño'),
    (8, 'salud'),
    (9, 'enfermo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Описание и состояния / Вкусы и предпочтения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Вкусы и предпочтения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Вкусы и предпочтения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'gusto'),
    (1, 'preferencia'),
    (2, 'interés'),
    (3, 'favorito'),
    (4, 'rico'),
    (5, 'dulce'),
    (6, 'salado'),
    (7, 'amargo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A1 / Описание и состояния / Противоположности и простые сравнения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Противоположности и простые сравнения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A1'
    AND cat.name = 'Описание и состояния'
    AND ws.title = 'Противоположности и простые сравнения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'igual'),
    (1, 'diferente'),
    (2, 'mejor'),
    (3, 'peor'),
    (4, 'mayor'),
    (5, 'menor'),
    (6, 'corto'),
    (7, 'largo'),
    (8, 'ancho'),
    (9, 'estrecho'),
    (10, 'lleno'),
    (11, 'vacío'),
    (12, 'fuerte'),
    (13, 'débil'),
    (14, 'importante')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Служебные слова и связность A2 / Местоименные формы и неопределённые слова
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Местоименные формы и неопределённые слова'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Местоименные формы и неопределённые слова'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'conmigo'),
    (1, 'contigo'),
    (2, 'consigo'),
    (3, 'cualquiera'),
    (4, 'ninguno'),
    (5, 'varios'),
    (6, 'ambos'),
    (7, 'cada'),
    (8, 'suficiente'),
    (9, 'cierto'),
    (10, 'propio'),
    (11, 'mismo'),
    (12, 'quienquiera'),
    (13, 'tal'),
    (14, 'semejante'),
    (15, 'bastantes'),
    (16, 'sendos'),
    (17, 'respectivo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Служебные слова и связность A2 / Предлоги и выражения времени, причины, цели
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Предлоги и выражения времени, причины, цели'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Предлоги и выражения времени, причины, цели'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'mediante'),
    (1, 'según'),
    (2, 'excepto'),
    (3, 'salvo'),
    (4, 'incluido'),
    (5, 'incluso'),
    (6, 'debido'),
    (7, 'causa'),
    (8, 'fuera'),
    (9, 'dentro'),
    (10, 'alrededor'),
    (11, 'junto'),
    (12, 'encima'),
    (13, 'debajo'),
    (14, 'tras'),
    (15, 'ante'),
    (16, 'final'),
    (17, 'principio')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Служебные слова и связность A2 / Союзы контраста, условия, последовательности
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Союзы контраста, условия, последовательности'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Союзы контраста, условия, последовательности'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sino'),
    (1, 'mas'),
    (2, 'siquiera'),
    (3, 'apenas'),
    (4, 'aparte'),
    (5, 'todavía'),
    (6, 'así'),
    (7, 'total'),
    (8, 'enseguida'),
    (9, 'finalmente'),
    (10, 'previamente'),
    (11, 'posteriormente'),
    (12, 'condición'),
    (13, 'caso'),
    (14, 'secuencia'),
    (15, 'contraste')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Служебные слова и связность A2 / Наречия вероятности, частотности, способа
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Наречия вероятности, частотности, способа'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Наречия вероятности, частотности, способа'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'quizá'),
    (1, 'quizás'),
    (2, 'probablemente'),
    (3, 'posiblemente'),
    (4, 'seguramente'),
    (5, 'normalmente'),
    (6, 'generalmente'),
    (7, 'especialmente'),
    (8, 'claramente'),
    (9, 'lentamente'),
    (10, 'rápidamente'),
    (11, 'fácilmente'),
    (12, 'difícilmente'),
    (13, 'casi'),
    (14, 'raramente'),
    (15, 'exactamente'),
    (16, 'aproximadamente'),
    (17, 'principalmente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Служебные слова и связность A2 / Маркеры диалога: уточнение, переспрос, реакции
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Маркеры диалога: уточнение, переспрос, реакции'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Служебные слова и связность A2'
    AND ws.title = 'Маркеры диалога: уточнение, переспрос, реакции'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'claro'),
    (1, 'vale'),
    (2, 'bueno'),
    (3, 'vaya'),
    (4, 'perdone'),
    (5, 'oiga'),
    (6, 'mire'),
    (7, 'repita'),
    (8, 'diga'),
    (9, 'perdona'),
    (10, 'disculpe'),
    (11, 'entiendo'),
    (12, 'exacto'),
    (13, 'correcto'),
    (14, 'efectivamente'),
    (15, 'seguro'),
    (16, 'genial'),
    (17, 'perfecto'),
    (18, 'acuerdo'),
    (19, 'duda')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Быт, дом, личные дела / Уборка, ремонт, бытовые предметы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Уборка, ремонт, бытовые предметы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Уборка, ремонт, бытовые предметы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'limpiar'),
    (1, 'barrer'),
    (2, 'fregar'),
    (3, 'lavar'),
    (4, 'planchar'),
    (5, 'ordenar'),
    (6, 'arreglar'),
    (7, 'reparar'),
    (8, 'cambiar'),
    (9, 'enchufe'),
    (10, 'bombilla'),
    (11, 'escoba'),
    (12, 'cubo'),
    (13, 'jabón'),
    (14, 'toalla'),
    (15, 'sábana'),
    (16, 'manta'),
    (17, 'herramienta')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Быт, дом, личные дела / Одежда, уход, покупки для дома
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Одежда, уход, покупки для дома'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Одежда, уход, покупки для дома'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'chaqueta'),
    (1, 'falda'),
    (2, 'jersey'),
    (3, 'calcetín'),
    (4, 'gorra'),
    (5, 'guante'),
    (6, 'cinturón'),
    (7, 'talla'),
    (8, 'color'),
    (9, 'crema'),
    (10, 'champú'),
    (11, 'cepillo'),
    (12, 'peine'),
    (13, 'espejo'),
    (14, 'perfume'),
    (15, 'detergente'),
    (16, 'cortina'),
    (17, 'alfombra')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Быт, дом, личные дела / Счета, аренда, коммуналка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Счета, аренда, коммуналка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Счета, аренда, коммуналка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'factura'),
    (1, 'tarifa'),
    (2, 'gasto'),
    (3, 'depósito'),
    (4, 'contrato'),
    (5, 'inquilino'),
    (6, 'propietario'),
    (7, 'electricidad'),
    (8, 'gas'),
    (9, 'suministro'),
    (10, 'calefacción'),
    (11, 'conexión'),
    (12, 'mensualidad'),
    (13, 'fianza'),
    (14, 'importe'),
    (15, 'pago'),
    (16, 'deuda'),
    (17, 'vencimiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Быт, дом, личные дела / Документы, анкеты, записи
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Документы, анкеты, записи'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Документы, анкеты, записи'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'solicitud'),
    (1, 'registro'),
    (2, 'inscripción'),
    (3, 'expediente'),
    (4, 'archivo'),
    (5, 'fotografía'),
    (6, 'identificación'),
    (7, 'licencia'),
    (8, 'resguardo'),
    (9, 'impreso'),
    (10, 'casilla'),
    (11, 'respuesta'),
    (12, 'pregunta'),
    (13, 'sección'),
    (14, 'página'),
    (15, 'trámite')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Быт, дом, личные дела / Проблемы и просьбы в быту
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Проблемы и просьбы в быту'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Быт, дом, личные дела'
    AND ws.title = 'Проблемы и просьбы в быту'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'problema'),
    (1, 'avería'),
    (2, 'ruido'),
    (3, 'humedad'),
    (4, 'fuga'),
    (5, 'mancha'),
    (6, 'rotura'),
    (7, 'urgente'),
    (8, 'molestar'),
    (9, 'avisar'),
    (10, 'llamar'),
    (11, 'solucionar'),
    (12, 'prestar'),
    (13, 'permitir'),
    (14, 'prohibir'),
    (15, 'queja'),
    (16, 'favor'),
    (17, 'ayuda'),
    (18, 'vecino'),
    (19, 'portero')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Профессии и рабочие роли
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Профессии и рабочие роли'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Профессии и рабочие роли'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'médico'),
    (1, 'profesor'),
    (2, 'ingeniero'),
    (3, 'abogado'),
    (4, 'enfermero'),
    (5, 'cocinero'),
    (6, 'dependiente'),
    (7, 'director'),
    (8, 'secretario'),
    (9, 'empleado'),
    (10, 'jefe'),
    (11, 'compañero'),
    (12, 'técnico'),
    (13, 'artista'),
    (14, 'periodista'),
    (15, 'chofer'),
    (16, 'recepcionista'),
    (17, 'limpiador'),
    (18, 'estudiante'),
    (19, 'trabajador')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Офисные действия и процессы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Офисные действия и процессы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Офисные действия и процессы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'organizar'),
    (1, 'preparar'),
    (2, 'presentar'),
    (3, 'enviar'),
    (4, 'revisar'),
    (5, 'firmar'),
    (6, 'imprimir'),
    (7, 'copiar'),
    (8, 'guardar'),
    (9, 'apuntar'),
    (10, 'anotar'),
    (11, 'confirmar'),
    (12, 'cancelar'),
    (13, 'reservar'),
    (14, 'planear'),
    (15, 'completar'),
    (16, 'entregar'),
    (17, 'recoger'),
    (18, 'reunirse'),
    (19, 'colaborar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Учёба, курсы, задания
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Учёба, курсы, задания'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Учёба, курсы, задания'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'curso'),
    (1, 'lección'),
    (2, 'ejercicio'),
    (3, 'examen'),
    (4, 'prueba'),
    (5, 'nota'),
    (6, 'tema'),
    (7, 'pizarra'),
    (8, 'libro'),
    (9, 'cuaderno'),
    (10, 'tarea'),
    (11, 'proyecto'),
    (12, 'práctica'),
    (13, 'explicación'),
    (14, 'ejemplo'),
    (15, 'corregir'),
    (16, 'aprender'),
    (17, 'enseñar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Компьютер, телефон, приложения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Компьютер, телефон, приложения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Компьютер, телефон, приложения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ordenador'),
    (1, 'pantalla'),
    (2, 'teclado'),
    (3, 'ratón'),
    (4, 'fichero'),
    (5, 'carpeta'),
    (6, 'programa'),
    (7, 'aplicación'),
    (8, 'contraseña'),
    (9, 'usuario'),
    (10, 'enlace'),
    (11, 'descargar'),
    (12, 'sincronizar'),
    (13, 'instalar'),
    (14, 'actualizar'),
    (15, 'borrar'),
    (16, 'reiniciar'),
    (17, 'transferir'),
    (18, 'conectar'),
    (19, 'desconectar'),
    (20, 'localizar'),
    (21, 'navegar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Переписка, сообщения, звонки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Переписка, сообщения, звонки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Переписка, сообщения, звонки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'texto'),
    (1, 'email'),
    (2, 'llamada'),
    (3, 'destinatario'),
    (4, 'chat'),
    (5, 'audio'),
    (6, 'vídeo'),
    (7, 'notificación'),
    (8, 'responder'),
    (9, 'contestar'),
    (10, 'marcar'),
    (11, 'colgar'),
    (12, 'sonar'),
    (13, 'grabar'),
    (14, 'adjuntar'),
    (15, 'reenviar'),
    (16, 'buzón'),
    (17, 'señal')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Работа, учёба, цифровая жизнь / Базовые IT- и интернет-слова
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Базовые IT- и интернет-слова'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Работа, учёба, цифровая жизнь'
    AND ws.title = 'Базовые IT- и интернет-слова'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'portal'),
    (1, 'interfaz'),
    (2, 'sitio'),
    (3, 'red'),
    (4, 'perfil'),
    (5, 'acceso'),
    (6, 'sesión'),
    (7, 'datos'),
    (8, 'privacidad'),
    (9, 'seguridad'),
    (10, 'buscador'),
    (11, 'servidor')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Здоровье, тело, спорт / Части тела A2
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Части тела A2'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Части тела A2'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cabeza'),
    (1, 'cara'),
    (2, 'ojo'),
    (3, 'oreja'),
    (4, 'nariz'),
    (5, 'boca'),
    (6, 'diente'),
    (7, 'cuello'),
    (8, 'hombro'),
    (9, 'brazo'),
    (10, 'mano'),
    (11, 'dedo'),
    (12, 'pierna'),
    (13, 'pie'),
    (14, 'espalda'),
    (15, 'estómago')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Здоровье, тело, спорт / Симптомы и простые болезни
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Симптомы и простые болезни'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Симптомы и простые болезни'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'fiebre'),
    (1, 'tos'),
    (2, 'gripe'),
    (3, 'resfriado'),
    (4, 'molestia'),
    (5, 'cansancio'),
    (6, 'mareo'),
    (7, 'náusea'),
    (8, 'herida'),
    (9, 'golpe'),
    (10, 'alergia'),
    (11, 'infección'),
    (12, 'enfermedad'),
    (13, 'síntoma'),
    (14, 'vomitar'),
    (15, 'toser'),
    (16, 'doler'),
    (17, 'sangrar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Здоровье, тело, спорт / Аптека и врач
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Аптека и врач'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Аптека и врач'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sanitario'),
    (1, 'doctor'),
    (2, 'auxiliar'),
    (3, 'paciente'),
    (4, 'consulta'),
    (5, 'receta'),
    (6, 'pastilla'),
    (7, 'jarabe'),
    (8, 'inyección'),
    (9, 'vacuna'),
    (10, 'tratamiento'),
    (11, 'análisis'),
    (12, 'turnero'),
    (13, 'urgencia'),
    (14, 'póliza'),
    (15, 'farmacéutico'),
    (16, 'administrar'),
    (17, 'curar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Здоровье, тело, спорт / Спорт и активность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Спорт и активность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Спорт и активность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'deporte'),
    (1, 'fútbol'),
    (2, 'baloncesto'),
    (3, 'natación'),
    (4, 'ciclismo'),
    (5, 'gimnasio'),
    (6, 'equipo'),
    (7, 'partido'),
    (8, 'jugar'),
    (9, 'entrenar'),
    (10, 'ganar'),
    (11, 'perder'),
    (12, 'campeón'),
    (13, 'entrenamiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Здоровье, тело, спорт / Привычки и самочувствие
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Привычки и самочувствие'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Здоровье, тело, спорт'
    AND ws.title = 'Привычки и самочувствие'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'costumbre'),
    (1, 'hábito'),
    (2, 'dieta'),
    (3, 'sano'),
    (4, 'saludable'),
    (5, 'activo'),
    (6, 'relajado'),
    (7, 'estresado'),
    (8, 'fumar'),
    (9, 'reposar'),
    (10, 'respirar'),
    (11, 'moverse'),
    (12, 'mejorar'),
    (13, 'cuidar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Транспорт A2
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Транспорт A2'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Транспорт A2'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'vehículo'),
    (1, 'coche'),
    (2, 'moto'),
    (3, 'bicicleta'),
    (4, 'camión'),
    (5, 'barco'),
    (6, 'avión'),
    (7, 'carretera'),
    (8, 'autopista'),
    (9, 'tráfico'),
    (10, 'semáforo'),
    (11, 'gasolina'),
    (12, 'aparcamiento'),
    (13, 'garaje'),
    (14, 'manejar'),
    (15, 'aparcar'),
    (16, 'frenar'),
    (17, 'arrancar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Поезд, аэропорт, билеты
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Поезд, аэропорт, билеты'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Поезд, аэропорт, билеты'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'vuelo'),
    (1, 'embarque'),
    (2, 'valija'),
    (3, 'bulto'),
    (4, 'documentación'),
    (5, 'control'),
    (6, 'mostrador'),
    (7, 'terminal'),
    (8, 'retraso'),
    (9, 'cancelación'),
    (10, 'butaca'),
    (11, 'vagón'),
    (12, 'vía'),
    (13, 'transbordo'),
    (14, 'llegada'),
    (15, 'salida'),
    (16, 'adquirir'),
    (17, 'facturar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Ориентирование в городе
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Ориентирование в городе'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Ориентирование в городе'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'barrio'),
    (1, 'zona'),
    (2, 'avenida'),
    (3, 'cruce'),
    (4, 'puente'),
    (5, 'rotonda'),
    (6, 'indicador'),
    (7, 'cartel'),
    (8, 'acera'),
    (9, 'entrada'),
    (10, 'desvío'),
    (11, 'subida'),
    (12, 'bajada'),
    (13, 'distancia'),
    (14, 'continuar'),
    (15, 'girar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Аренда, гостиница, жильё
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Аренда, гостиница, жильё'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Аренда, гостиница, жильё'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'reservación'),
    (1, 'cuarto'),
    (2, 'conserje'),
    (3, 'credencial'),
    (4, 'visitante'),
    (5, 'pernocta'),
    (6, 'colchón'),
    (7, 'aseo'),
    (8, 'amenidad'),
    (9, 'desayuno'),
    (10, 'limpieza'),
    (11, 'ascensor'),
    (12, 'balcón'),
    (13, 'arrendar'),
    (14, 'mudarse'),
    (15, 'estancia'),
    (16, 'disponible'),
    (17, 'reservado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Места и учреждения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Места и учреждения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Места и учреждения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ayuntamiento'),
    (1, 'juzgado'),
    (2, 'colegio'),
    (3, 'universidad'),
    (4, 'supermercado'),
    (5, 'instituto'),
    (6, 'polideportivo'),
    (7, 'piscina'),
    (8, 'gasolinera'),
    (9, 'taller'),
    (10, 'peluquería'),
    (11, 'panadería'),
    (12, 'carnicería'),
    (13, 'verdulería'),
    (14, 'cafetería'),
    (15, 'discoteca')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Путешествия, город, жильё / Безопасность, потеря, поломка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Безопасность, потеря, поломка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Путешествия, город, жильё'
    AND ws.title = 'Безопасность, потеря, поломка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'peligro'),
    (1, 'robo'),
    (2, 'ladrón'),
    (3, 'policía'),
    (4, 'emergencia'),
    (5, 'accidente'),
    (6, 'daño'),
    (7, 'pérdida'),
    (8, 'extraviar'),
    (9, 'romper'),
    (10, 'descuidar'),
    (11, 'denunciar'),
    (12, 'recuperar'),
    (13, 'cuidado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Общение, мнения, эмоции / Простые мнения и оценки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Простые мнения и оценки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Простые мнения и оценки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'opinión'),
    (1, 'idea'),
    (2, 'razón'),
    (3, 'muestra'),
    (4, 'verdad'),
    (5, 'mentira'),
    (6, 'interesante'),
    (7, 'aburrido'),
    (8, 'útil'),
    (9, 'inútil'),
    (10, 'normal'),
    (11, 'extraño'),
    (12, 'especial'),
    (13, 'necesario'),
    (14, 'posible'),
    (15, 'imposible'),
    (16, 'valer'),
    (17, 'importar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Общение, мнения, эмоции / Согласие, несогласие, сомнение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Согласие, несогласие, сомнение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Согласие, несогласие, сомнение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'aceptar'),
    (1, 'rechazar'),
    (2, 'negar'),
    (3, 'dudar'),
    (4, 'ratificar'),
    (5, 'acordar'),
    (6, 'discutir'),
    (7, 'opinar'),
    (8, 'convencido'),
    (9, 'inseguro'),
    (10, 'talvez'),
    (11, 'depende'),
    (12, 'veraz'),
    (13, 'falso'),
    (14, 'conformidad'),
    (15, 'desacuerdo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Общение, мнения, эмоции / Эмоции и характер
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Эмоции и характер'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Эмоции и характер'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'carácter'),
    (1, 'ánimo'),
    (2, 'alegre'),
    (3, 'enfadado'),
    (4, 'preocupado'),
    (5, 'sorprendido'),
    (6, 'asustado'),
    (7, 'contento'),
    (8, 'orgulloso'),
    (9, 'tímido'),
    (10, 'valiente'),
    (11, 'tolerante'),
    (12, 'impaciente'),
    (13, 'sensible'),
    (14, 'cordial'),
    (15, 'egoísta'),
    (16, 'generoso'),
    (17, 'formal')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Общение, мнения, эмоции / Отношения и социальные ситуации
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Отношения и социальные ситуации'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Отношения и социальные ситуации'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'relación'),
    (1, 'amistad'),
    (2, 'confianza'),
    (3, 'respeto'),
    (4, 'fiesta'),
    (5, 'visita'),
    (6, 'invitado'),
    (7, 'anfitrión'),
    (8, 'grupo'),
    (9, 'conocido'),
    (10, 'compañía'),
    (11, 'conocer'),
    (12, 'relacionar'),
    (13, 'invitar'),
    (14, 'saludar'),
    (15, 'despedirse')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Общение, мнения, эмоции / Просьбы, извинения, договорённости
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Просьбы, извинения, договорённости'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Общение, мнения, эмоции'
    AND ws.title = 'Просьбы, извинения, договорённости'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'prometer'),
    (1, 'ofrecer'),
    (2, 'proponer'),
    (3, 'concertar'),
    (4, 'citarse'),
    (5, 'admitir'),
    (6, 'anular'),
    (7, 'apologizar'),
    (8, 'agradecer'),
    (9, 'perdonar'),
    (10, 'rogar'),
    (11, 'exigir'),
    (12, 'sugerir'),
    (13, 'insistir'),
    (14, 'notificar'),
    (15, 'verificar'),
    (16, 'retrasar'),
    (17, 'aplazar'),
    (18, 'resolver'),
    (19, 'excusa'),
    (20, 'autorización'),
    (21, 'apoyo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Медиа, культура, досуг / Фильмы, музыка, книги
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Фильмы, музыка, книги'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Фильмы, музыка, книги'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'película'),
    (1, 'serie'),
    (2, 'canción'),
    (3, 'música'),
    (4, 'concierto'),
    (5, 'cantante'),
    (6, 'actor'),
    (7, 'actriz'),
    (8, 'realizador'),
    (9, 'novela'),
    (10, 'cuento'),
    (11, 'poema'),
    (12, 'autor'),
    (13, 'lector'),
    (14, 'escena'),
    (15, 'boleto')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Медиа, культура, досуг / Хобби и свободное время
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Хобби и свободное время'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Хобби и свободное время'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'hobby'),
    (1, 'afición'),
    (2, 'pintar'),
    (3, 'dibujar'),
    (4, 'bailar'),
    (5, 'cantar'),
    (6, 'tocar'),
    (7, 'leer'),
    (8, 'escribir'),
    (9, 'fotografiar'),
    (10, 'coleccionar'),
    (11, 'pasear'),
    (12, 'jardinería'),
    (13, 'manualidad'),
    (14, 'juego'),
    (15, 'ocio')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Медиа, культура, досуг / События, праздники, приглашения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'События, праздники, приглашения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'События, праздники, приглашения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'evento'),
    (1, 'celebración'),
    (2, 'cumpleaños'),
    (3, 'boda'),
    (4, 'aniversario'),
    (5, 'invitación'),
    (6, 'regalo'),
    (7, 'sorpresa'),
    (8, 'brindis'),
    (9, 'decoración'),
    (10, 'pase'),
    (11, 'agenda'),
    (12, 'asistir'),
    (13, 'celebrar'),
    (14, 'coordinar'),
    (15, 'convocar'),
    (16, 'aceptación'),
    (17, 'declinar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Медиа, культура, досуг / Интернет-контент и соцсети
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Интернет-контент и соцсети'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Интернет-контент и соцсети'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'publicación'),
    (1, 'comentario'),
    (2, 'foto'),
    (3, 'clip'),
    (4, 'canal'),
    (5, 'plataforma'),
    (6, 'seguidor'),
    (7, 'noticia'),
    (8, 'titular'),
    (9, 'compartir'),
    (10, 'publicar'),
    (11, 'comentar'),
    (12, 'registrarse'),
    (13, 'bloquear'),
    (14, 'etiquetar'),
    (15, 'afiliarse')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Медиа, культура, досуг / Культурные места
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Культурные места'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Медиа, культура, досуг'
    AND ws.title = 'Культурные места'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'galería'),
    (1, 'exposición'),
    (2, 'monumento'),
    (3, 'castillo'),
    (4, 'palacio'),
    (5, 'mediateca'),
    (6, 'sala'),
    (7, 'auditorio'),
    (8, 'escenario'),
    (9, 'obra'),
    (10, 'festival'),
    (11, 'ingreso'),
    (12, 'guía'),
    (13, 'recorrido')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Природа, погода, животные, еда / Природа и ландшафт
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Природа и ландшафт'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Природа и ландшафт'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'naturaleza'),
    (1, 'campo'),
    (2, 'bosque'),
    (3, 'montaña'),
    (4, 'río'),
    (5, 'lago'),
    (6, 'mar'),
    (7, 'playa'),
    (8, 'isla'),
    (9, 'valle'),
    (10, 'piedra'),
    (11, 'arena'),
    (12, 'cielo'),
    (13, 'tierra')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Природа, погода, животные, еда / Животные и растения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Животные и растения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Животные и растения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'animal'),
    (1, 'perro'),
    (2, 'gato'),
    (3, 'caballo'),
    (4, 'pájaro'),
    (5, 'pez'),
    (6, 'árbol'),
    (7, 'flor'),
    (8, 'planta'),
    (9, 'hoja'),
    (10, 'hierba'),
    (11, 'semilla')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Природа, погода, животные, еда / Погода и климат A2
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Погода и климат A2'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Погода и климат A2'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'clima'),
    (1, 'temperatura'),
    (2, 'grado'),
    (3, 'tormenta'),
    (4, 'niebla'),
    (5, 'hielo'),
    (6, 'llover'),
    (7, 'nevar'),
    (8, 'soplar'),
    (9, 'nublado'),
    (10, 'soleado'),
    (11, 'seco'),
    (12, 'húmedo'),
    (13, 'templado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- A2 / Природа, погода, животные, еда / Еда, кухня, рецепты A2
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Еда, кухня, рецепты A2'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень A2'
    AND cat.name = 'Природа, погода, животные, еда'
    AND ws.title = 'Еда, кухня, рецепты A2'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ingrediente'),
    (1, 'preparación'),
    (2, 'cuchillo'),
    (3, 'tenedor'),
    (4, 'cuchara'),
    (5, 'olla'),
    (6, 'sartén'),
    (7, 'horno'),
    (8, 'freír'),
    (9, 'hervir'),
    (10, 'asar'),
    (11, 'mezclar'),
    (12, 'cortar'),
    (13, 'pelar'),
    (14, 'añadir'),
    (15, 'degustar'),
    (16, 'sabor'),
    (17, 'picante'),
    (18, 'caliente'),
    (19, 'fresco')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Вводные слова и структура мысли
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Вводные слова и структура мысли'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Вводные слова и структура мысли'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'asimismo'),
    (1, 'inicialmente'),
    (2, 'conclusivamente'),
    (3, 'concretamente'),
    (4, 'brevemente'),
    (5, 'sinceramente'),
    (6, 'personalmente'),
    (7, 'fehacientemente'),
    (8, 'básicamente'),
    (9, 'esquemáticamente'),
    (10, 'teóricamente'),
    (11, 'empíricamente'),
    (12, 'analíticamente'),
    (13, 'consecuentemente'),
    (14, 'paralelamente'),
    (15, 'globalmente'),
    (16, 'parcialmente'),
    (17, 'esencialmente'),
    (18, 'ciertamente'),
    (19, 'recapitulando'),
    (20, 'resumiendo'),
    (21, 'planteamiento'),
    (22, 'enfoque'),
    (23, 'argumento'),
    (24, 'apartado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Причина, следствие, цель, условие
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Причина, следствие, цель, условие'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Причина, следствие, цель, условие'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'causalidad'),
    (1, 'motivo'),
    (2, 'consecuencia'),
    (3, 'efecto'),
    (4, 'finalidad'),
    (5, 'propósito'),
    (6, 'condicionante'),
    (7, 'requisito'),
    (8, 'supuesto'),
    (9, 'dependencia'),
    (10, 'determinante'),
    (11, 'detonante'),
    (12, 'incidencia'),
    (13, 'repercusión'),
    (14, 'derivación'),
    (15, 'desencadenante'),
    (16, 'provocar'),
    (17, 'generar'),
    (18, 'ocasionar'),
    (19, 'facilitar'),
    (20, 'impedir'),
    (21, 'contribuir'),
    (22, 'justificar'),
    (23, 'determinar'),
    (24, 'condicionar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Уступка, контраст, исключение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Уступка, контраст, исключение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Уступка, контраст, исключение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'contraposición'),
    (1, 'oposición'),
    (2, 'excepción'),
    (3, 'objeción'),
    (4, 'salvedad'),
    (5, 'restricción'),
    (6, 'matiz'),
    (7, 'límite'),
    (8, 'alternativa'),
    (9, 'discrepancia'),
    (10, 'concesión'),
    (11, 'divergencia'),
    (12, 'contrario'),
    (13, 'contrapunto'),
    (14, 'contrapeso'),
    (15, 'excluir'),
    (16, 'exceptuar'),
    (17, 'distinguir'),
    (18, 'relativizar'),
    (19, 'puntualizar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Степень уверенности и вероятности
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Степень уверенности и вероятности'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Степень уверенности и вероятности'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'certeza'),
    (1, 'incertidumbre'),
    (2, 'probabilidad'),
    (3, 'posibilidad'),
    (4, 'hipótesis'),
    (5, 'sospecha'),
    (6, 'evidencia'),
    (7, 'indicio'),
    (8, 'conjetura'),
    (9, 'previsión'),
    (10, 'estimación'),
    (11, 'cálculo'),
    (12, 'convicción'),
    (13, 'vacilación'),
    (14, 'plausible'),
    (15, 'verosímil'),
    (16, 'incierto'),
    (17, 'improbable'),
    (18, 'presumible'),
    (19, 'dudoso')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Временные отношения и последовательность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Временные отношения и последовательность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Временные отношения и последовательность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'anterioridad'),
    (1, 'posterioridad'),
    (2, 'simultaneidad'),
    (3, 'sucesión'),
    (4, 'precedencia'),
    (5, 'intervalo'),
    (6, 'periodo'),
    (7, 'plazo'),
    (8, 'continuidad'),
    (9, 'interrupción'),
    (10, 'pausa'),
    (11, 'reanudación'),
    (12, 'inicio'),
    (13, 'cierre'),
    (14, 'transcurso'),
    (15, 'preceder'),
    (16, 'suceder'),
    (17, 'reanudarse'),
    (18, 'prolongarse'),
    (19, 'culminar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Связность, грамматика речи, дискурс / Пересказ, уточнение, примеры
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Пересказ, уточнение, примеры'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Связность, грамматика речи, дискурс'
    AND ws.title = 'Пересказ, уточнение, примеры'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'resumir'),
    (1, 'sintetizar'),
    (2, 'aclarar'),
    (3, 'precisar'),
    (4, 'matizar'),
    (5, 'detallar'),
    (6, 'reformular'),
    (7, 'parafrasear'),
    (8, 'citar'),
    (9, 'mencionar'),
    (10, 'ejemplificar'),
    (11, 'ilustrar'),
    (12, 'señalar'),
    (13, 'indicar'),
    (14, 'destacar'),
    (15, 'agregar'),
    (16, 'omitir'),
    (17, 'narrar'),
    (18, 'relatar'),
    (19, 'describir'),
    (20, 'interpretar'),
    (21, 'apostillar'),
    (22, 'referencia'),
    (23, 'fuente'),
    (24, 'versión'),
    (25, 'esclarecimiento'),
    (26, 'ilustración'),
    (27, 'antecedente'),
    (28, 'aclaración'),
    (29, 'precisión')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Вакансии, резюме, собеседование
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Вакансии, резюме, собеседование'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Вакансии, резюме, собеседование'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'vacante'),
    (1, 'candidatura'),
    (2, 'currículum'),
    (3, 'entrevista'),
    (4, 'salario'),
    (5, 'puesto'),
    (6, 'selección'),
    (7, 'candidato'),
    (8, 'reclutador'),
    (9, 'disponibilidad'),
    (10, 'recomendación'),
    (11, 'trayectoria'),
    (12, 'portafolio'),
    (13, 'experiencia'),
    (14, 'semblanza'),
    (15, 'aptitud'),
    (16, 'contratación'),
    (17, 'jornada'),
    (18, 'postular'),
    (19, 'negociar'),
    (20, 'contratar'),
    (21, 'ascender'),
    (22, 'incorporarse'),
    (23, 'renunciar'),
    (24, 'preselección')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Обязанности, навыки, условия
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Обязанности, навыки, условия'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Обязанности, навыки, условия'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'responsabilidad'),
    (1, 'función'),
    (2, 'competencia'),
    (3, 'destreza'),
    (4, 'exigencia'),
    (5, 'sueldo'),
    (6, 'beneficio'),
    (7, 'formación'),
    (8, 'puntualidad'),
    (9, 'iniciativa'),
    (10, 'liderazgo'),
    (11, 'productividad'),
    (12, 'autonomía'),
    (13, 'flexibilidad'),
    (14, 'compromiso'),
    (15, 'diligente'),
    (16, 'eficaz'),
    (17, 'organizado'),
    (18, 'cumplir'),
    (19, 'armonizar'),
    (20, 'dominar'),
    (21, 'perfeccionar'),
    (22, 'supervisar'),
    (23, 'gestionar'),
    (24, 'capacitar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Офисные процессы и коммуникация
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Офисные процессы и коммуникация'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Офисные процессы и коммуникация'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'informe'),
    (1, 'acta'),
    (2, 'orden'),
    (3, 'encuentro'),
    (4, 'comunicado'),
    (5, 'aviso'),
    (6, 'propuesta'),
    (7, 'interpelación'),
    (8, 'contestación'),
    (9, 'seguimiento'),
    (10, 'aprobación'),
    (11, 'revisión'),
    (12, 'petición'),
    (13, 'comunicación'),
    (14, 'convocatoria'),
    (15, 'remitente'),
    (16, 'receptor'),
    (17, 'solicitar'),
    (18, 'replicar'),
    (19, 'informar'),
    (20, 'circular'),
    (21, 'trasladar'),
    (22, 'remitir'),
    (23, 'anexar'),
    (24, 'archivar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Проекты, сроки, задачи
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Проекты, сроки, задачи'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Проекты, сроки, задачи'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'plan'),
    (1, 'cometido'),
    (2, 'cadencia'),
    (3, 'entrega'),
    (4, 'prioridad'),
    (5, 'avance'),
    (6, 'demora'),
    (7, 'recurso'),
    (8, 'presupuesto'),
    (9, 'responsable'),
    (10, 'cronograma'),
    (11, 'planificación'),
    (12, 'objetivo'),
    (13, 'hito'),
    (14, 'etapa'),
    (15, 'bloqueo'),
    (16, 'cómputo'),
    (17, 'asignación'),
    (18, 'interdependencia'),
    (19, 'pendiente'),
    (20, 'asignar'),
    (21, 'finalizar'),
    (22, 'comprobar'),
    (23, 'lanzar'),
    (24, 'clausurar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Ошибки, конфликты, решения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Ошибки, конфликты, решения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Ошибки, конфликты, решения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'fallo'),
    (1, 'conflicto'),
    (2, 'tensión'),
    (3, 'reclamación'),
    (4, 'crítica'),
    (5, 'incidente'),
    (6, 'fricción'),
    (7, 'malentendido'),
    (8, 'solución'),
    (9, 'secuelas'),
    (10, 'reparación'),
    (11, 'mediación'),
    (12, 'disculparse'),
    (13, 'mediar'),
    (14, 'rectificar'),
    (15, 'solventar'),
    (16, 'esclarecer'),
    (17, 'asumir'),
    (18, 'evitar'),
    (19, 'plantear')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Работа и карьера / Удалённая работа и фриланс
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Удалённая работа и фриланс'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Работа и карьера'
    AND ws.title = 'Удалённая работа и фриланс'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'teletrabajo'),
    (1, 'remoto'),
    (2, 'autónomo'),
    (3, 'freelance'),
    (4, 'encargo'),
    (5, 'videollamada'),
    (6, 'portátil'),
    (7, 'colaboración'),
    (8, 'facturación'),
    (9, 'asincrónico'),
    (10, 'briefing'),
    (11, 'anticipo'),
    (12, 'entregable'),
    (13, 'externalizar'),
    (14, 'subcontratar'),
    (15, 'telemático'),
    (16, 'coworking'),
    (17, 'consultoría'),
    (18, 'honorario'),
    (19, 'proveedor'),
    (20, 'autónomamente'),
    (21, 'deslocalizado'),
    (22, 'descentralizado'),
    (23, 'subcontratación'),
    (24, 'intermediación'),
    (25, 'facturable'),
    (26, 'retainer'),
    (27, 'cotización'),
    (28, 'consultor'),
    (29, 'nomadismo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Образование и саморазвитие / Школа, университет, предметы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Школа, университет, предметы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Школа, университет, предметы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'asignatura'),
    (1, 'materia'),
    (2, 'facultad'),
    (3, 'carrera'),
    (4, 'campus'),
    (5, 'semestre'),
    (6, 'aula'),
    (7, 'alumno'),
    (8, 'historia'),
    (9, 'literatura'),
    (10, 'matemática'),
    (11, 'ciencia'),
    (12, 'física'),
    (13, 'química'),
    (14, 'biología'),
    (15, 'geografía'),
    (16, 'filosofía'),
    (17, 'economía'),
    (18, 'psicología'),
    (19, 'sociología')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Образование и саморазвитие / Курсы, экзамены, оценка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Курсы, экзамены, оценка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Курсы, экзамены, оценка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'evaluación'),
    (1, 'aprobado'),
    (2, 'suspenso'),
    (3, 'diploma'),
    (4, 'matrícula'),
    (5, 'corrección'),
    (6, 'calificación'),
    (7, 'recuperación'),
    (8, 'repasar'),
    (9, 'aprobar'),
    (10, 'suspender'),
    (11, 'evaluar'),
    (12, 'calificar'),
    (13, 'baremo'),
    (14, 'rúbrica'),
    (15, 'temario'),
    (16, 'tutoría'),
    (17, 'simulacro'),
    (18, 'oral'),
    (19, 'escrito')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Образование и саморазвитие / Навыки и обучение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Навыки и обучение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Навыки и обучение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'aprendizaje'),
    (1, 'conocimiento'),
    (2, 'método'),
    (3, 'memoria'),
    (4, 'atención'),
    (5, 'concentración'),
    (6, 'disciplina'),
    (7, 'curiosidad'),
    (8, 'esfuerzo'),
    (9, 'progreso'),
    (10, 'estrategia'),
    (11, 'rutina'),
    (12, 'asimilar'),
    (13, 'memorizar'),
    (14, 'ejercitar'),
    (15, 'capacitarse'),
    (16, 'perfeccionarse'),
    (17, 'reforzar'),
    (18, 'habilidad'),
    (19, 'autodidacta')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Образование и саморазвитие / Чтение, исследование, источники
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Чтение, исследование, источники'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Чтение, исследование, источники'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'lectura'),
    (1, 'investigación'),
    (2, 'artículo'),
    (3, 'capítulo'),
    (4, 'índice'),
    (5, 'bibliografía'),
    (6, 'resumen'),
    (7, 'consultar'),
    (8, 'comparar'),
    (9, 'subrayar'),
    (10, 'clasificar'),
    (11, 'hemeroteca'),
    (12, 'monografía'),
    (13, 'tesis'),
    (14, 'extracto'),
    (15, 'glosario'),
    (16, 'reseña'),
    (17, 'ficha'),
    (18, 'catalogar'),
    (19, 'documentarse')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Образование и саморазвитие / Цели, прогресс, результаты
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Цели, прогресс, результаты'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Образование и саморазвитие'
    AND ws.title = 'Цели, прогресс, результаты'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'meta'),
    (1, 'logro'),
    (2, 'éxito'),
    (3, 'fracaso'),
    (4, 'constancia'),
    (5, 'rendimiento'),
    (6, 'resultado'),
    (7, 'superación'),
    (8, 'alcanzar'),
    (9, 'progresar'),
    (10, 'medir'),
    (11, 'perseverar'),
    (12, 'evolución'),
    (13, 'cumplimiento'),
    (14, 'balance'),
    (15, 'ambición'),
    (16, 'aspiración'),
    (17, 'desempeño'),
    (18, 'madurez'),
    (19, 'culminación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Госучреждения и процедуры
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Госучреждения и процедуры'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Госучреждения и процедуры'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'delegación'),
    (1, 'negociado'),
    (2, 'administrativo'),
    (3, 'acreditación'),
    (4, 'tasa'),
    (5, 'comparecencia'),
    (6, 'compulsar'),
    (7, 'validación'),
    (8, 'fedatario'),
    (9, 'diligencia'),
    (10, 'subsanación'),
    (11, 'personación'),
    (12, 'certificación'),
    (13, 'registrador'),
    (14, 'gestoría'),
    (15, 'burocracia'),
    (16, 'funcionariado'),
    (17, 'tramitación'),
    (18, 'folio'),
    (19, 'legajo'),
    (20, 'providencia'),
    (21, 'instancia'),
    (22, 'emplazamiento'),
    (23, 'ventanilla'),
    (24, 'administración')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Виза, резиденция, регистрация
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Виза, резиденция, регистрация'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Виза, резиденция, регистрация'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'extranjería'),
    (1, 'repatriación'),
    (2, 'expatriado'),
    (3, 'migrante'),
    (4, 'solicitante'),
    (5, 'empadronarse'),
    (6, 'naturalizarse'),
    (7, 'arraigo'),
    (8, 'apátrida'),
    (9, 'refugiarse'),
    (10, 'temporal'),
    (11, 'permanente'),
    (12, 'reagrupación'),
    (13, 'consular'),
    (14, 'biométrico'),
    (15, 'prórroga'),
    (16, 'inadmisión'),
    (17, 'deportación'),
    (18, 'nacionalización'),
    (19, 'regularización'),
    (20, 'renovable'),
    (21, 'caducidad'),
    (22, 'padrón'),
    (23, 'frontera'),
    (24, 'consulado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Банк, страховка, налоги: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Банк, страховка, налоги: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Банк, страховка, налоги: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'gravamen'),
    (1, 'fiscalidad'),
    (2, 'contribuyente'),
    (3, 'retención'),
    (4, 'liquidación'),
    (5, 'inspección'),
    (6, 'exención'),
    (7, 'asegurado'),
    (8, 'siniestro'),
    (9, 'tributación'),
    (10, 'deducible'),
    (11, 'imponible'),
    (12, 'recaudación'),
    (13, 'hacienda'),
    (14, 'copago'),
    (15, 'franquicia'),
    (16, 'solvencia'),
    (17, 'morosidad'),
    (18, 'recargo'),
    (19, 'deducción'),
    (20, 'aseguradora'),
    (21, 'prima'),
    (22, 'cobertura'),
    (23, 'cuota'),
    (24, 'saldo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Аренда, договор, права, обязанности
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Аренда, договор, права, обязанности'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Аренда, договор, права, обязанности'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'arrendador'),
    (1, 'arrendatario'),
    (2, 'subarriendo'),
    (3, 'usufructo'),
    (4, 'cesión'),
    (5, 'aval'),
    (6, 'desperfecto'),
    (7, 'inventario'),
    (8, 'rescisión'),
    (9, 'vigencia'),
    (10, 'arrendamiento'),
    (11, 'cláusula'),
    (12, 'habitabilidad'),
    (13, 'caución'),
    (14, 'moroso'),
    (15, 'desahucio'),
    (16, 'inmueble'),
    (17, 'ocupante'),
    (18, 'notarial'),
    (19, 'escriturar'),
    (20, 'usufructuario'),
    (21, 'canon'),
    (22, 'renta'),
    (23, 'posesión'),
    (24, 'copropiedad')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Письма, заявления, формы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Письма, заявления, формы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Письма, заявления, формы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'membrete'),
    (1, 'acuse'),
    (2, 'enmienda'),
    (3, 'alegación'),
    (4, 'borrador'),
    (5, 'encabezado'),
    (6, 'diligenciar'),
    (7, 'rubricar'),
    (8, 'protocolizar'),
    (9, 'registrar'),
    (10, 'remisión'),
    (11, 'subsanar'),
    (12, 'transcribir'),
    (13, 'certificar'),
    (14, 'requerimiento'),
    (15, 'adjunto'),
    (16, 'asunto'),
    (17, 'carta'),
    (18, 'suplico'),
    (19, 'expongo'),
    (20, 'notificable'),
    (21, 'suscripción'),
    (22, 'rubricado'),
    (23, 'sellado'),
    (24, 'foliación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Администрация, переезд, документы / Чрезвычайные и юридические ситуации
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Чрезвычайные и юридические ситуации'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Администрация, переезд, документы'
    AND ws.title = 'Чрезвычайные и юридические ситуации'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'imputado'),
    (1, 'infracción'),
    (2, 'querella'),
    (3, 'custodia'),
    (4, 'auxilio'),
    (5, 'sanción'),
    (6, 'agresión'),
    (7, 'comparecer'),
    (8, 'socorrer'),
    (9, 'delito'),
    (10, 'denuncia'),
    (11, 'testigo'),
    (12, 'víctima'),
    (13, 'amenaza'),
    (14, 'flagrancia'),
    (15, 'hurto'),
    (16, 'estafa'),
    (17, 'coacción'),
    (18, 'atestado'),
    (19, 'detención'),
    (20, 'arresto'),
    (21, 'penal'),
    (22, 'civil'),
    (23, 'juicio'),
    (24, 'peritaje')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Здоровье и бытовые проблемы / Врач, клиника, запись
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Врач, клиника, запись'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Врач, клиника, запись'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'clínica'),
    (1, 'especialista'),
    (2, 'diagnóstico'),
    (3, 'historial'),
    (4, 'ambulatorio'),
    (5, 'auscultar'),
    (6, 'diagnosticar'),
    (7, 'ingresar'),
    (8, 'triaje'),
    (9, 'derivar'),
    (10, 'revisarse'),
    (11, 'atender'),
    (12, 'examinar'),
    (13, 'auscultación'),
    (14, 'volante'),
    (15, 'previa'),
    (16, 'cabecera'),
    (17, 'pediatra'),
    (18, 'traumatólogo'),
    (19, 'dermatólogo'),
    (20, 'analítica'),
    (21, 'enfermería'),
    (22, 'cardiólogo'),
    (23, 'ginecólogo'),
    (24, 'otorrino')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Здоровье и бытовые проблемы / Симптомы, травмы, боль
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Симптомы, травмы, боль'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Симптомы, травмы, боль'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'lesión'),
    (1, 'fractura'),
    (2, 'quemadura'),
    (3, 'hinchazón'),
    (4, 'inflamación'),
    (5, 'sangrado'),
    (6, 'esguince'),
    (7, 'hematoma'),
    (8, 'calambre'),
    (9, 'jaqueca'),
    (10, 'cicatriz'),
    (11, 'entumecimiento'),
    (12, 'punzada'),
    (13, 'empeorar'),
    (14, 'ampolla'),
    (15, 'moretón'),
    (16, 'picor'),
    (17, 'rigidez'),
    (18, 'contractura'),
    (19, 'luxación'),
    (20, 'corte'),
    (21, 'rasguño'),
    (22, 'desmayo'),
    (23, 'escalofrío'),
    (24, 'palpitación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Здоровье и бытовые проблемы / Лечение, лекарства, рекомендации
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Лечение, лекарства, рекомендации'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Лечение, лекарства, рекомендации'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'medicamento'),
    (1, 'dosis'),
    (2, 'antibiótico'),
    (3, 'pomada'),
    (4, 'reposo'),
    (5, 'prospecto'),
    (6, 'pauta'),
    (7, 'aliviar'),
    (8, 'prevenir'),
    (9, 'aplicar'),
    (10, 'recetar'),
    (11, 'dosificar'),
    (12, 'desinfectar'),
    (13, 'analgésico'),
    (14, 'antiinflamatorio'),
    (15, 'comprimido'),
    (16, 'cápsula'),
    (17, 'vendaje'),
    (18, 'hidratar'),
    (19, 'inhalador'),
    (20, 'termómetro'),
    (21, 'contraindicación'),
    (22, 'posología'),
    (23, 'suero'),
    (24, 'apósito')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Здоровье и бытовые проблемы / Ремонт, поломки, сервис
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Ремонт, поломки, сервис'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Ремонт, поломки, сервис'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'mantenimiento'),
    (1, 'instalación'),
    (2, 'garantía'),
    (3, 'pieza'),
    (4, 'arreglo'),
    (5, 'sustituir'),
    (6, 'ajustar'),
    (7, 'fontanero'),
    (8, 'electricista'),
    (9, 'albañil'),
    (10, 'repuesto'),
    (11, 'taladro'),
    (12, 'tornillo'),
    (13, 'tuerca'),
    (14, 'grifo'),
    (15, 'persiana'),
    (16, 'cerradura'),
    (17, 'cañería'),
    (18, 'desagüe'),
    (19, 'interruptor')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Здоровье и бытовые проблемы / Бытовые риски и безопасность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Бытовые риски и безопасность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Здоровье и бытовые проблемы'
    AND ws.title = 'Бытовые риски и безопасность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'incendio'),
    (1, 'caída'),
    (2, 'alarma'),
    (3, 'vigilancia'),
    (4, 'prevención'),
    (5, 'extintor'),
    (6, 'evacuación'),
    (7, 'cortocircuito'),
    (8, 'escape'),
    (9, 'intoxicación'),
    (10, 'resbalón'),
    (11, 'proteger'),
    (12, 'vigilar'),
    (13, 'escapar'),
    (14, 'asegurar'),
    (15, 'detector'),
    (16, 'botiquín'),
    (17, 'barandilla'),
    (18, 'antideslizante'),
    (19, 'mascarilla'),
    (20, 'precaución'),
    (21, 'cerrojo'),
    (22, 'ventilación'),
    (23, 'extinción'),
    (24, 'desinfectante')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Отношения, характер, психология / Характер и личные качества
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Характер и личные качества'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Характер и личные качества'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'personalidad'),
    (1, 'cualidad'),
    (2, 'defecto'),
    (3, 'honesto'),
    (4, 'creativo'),
    (5, 'sincero'),
    (6, 'impulsivo'),
    (7, 'prudente'),
    (8, 'sociable'),
    (9, 'terco'),
    (10, 'humilde'),
    (11, 'optimista'),
    (12, 'pesimista'),
    (13, 'leal'),
    (14, 'maduro'),
    (15, 'testarudo'),
    (16, 'empático'),
    (17, 'desconfiado'),
    (18, 'ambicioso'),
    (19, 'reflexivo'),
    (20, 'atrevido'),
    (21, 'introvertido'),
    (22, 'extrovertido'),
    (23, 'perfeccionista'),
    (24, 'rencoroso')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Отношения, характер, психология / Эмоции и состояния B1
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Эмоции и состояния B1'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Эмоции и состояния B1'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ansiedad'),
    (1, 'calma'),
    (2, 'rabia'),
    (3, 'vergüenza'),
    (4, 'orgullo'),
    (5, 'alivio'),
    (6, 'entusiasmo'),
    (7, 'decepción'),
    (8, 'esperanza'),
    (9, 'frustración'),
    (10, 'emoción'),
    (11, 'angustia'),
    (12, 'irritación'),
    (13, 'ternura'),
    (14, 'nostalgia'),
    (15, 'celos'),
    (16, 'euforia'),
    (17, 'alegrarse'),
    (18, 'preocuparse'),
    (19, 'tranquilizarse'),
    (20, 'agobio'),
    (21, 'serenidad'),
    (22, 'inquietud'),
    (23, 'desánimo'),
    (24, 'satisfacción')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Отношения, характер, психология / Дружба, семья, отношения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Дружба, семья, отношения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Дружба, семья, отношения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cariño'),
    (1, 'discusión'),
    (2, 'vínculo'),
    (3, 'convivencia'),
    (4, 'intimidad'),
    (5, 'lealtad'),
    (6, 'reconciliación'),
    (7, 'complicidad'),
    (8, 'cercanía'),
    (9, 'afecto'),
    (10, 'ruptura'),
    (11, 'acompañar'),
    (12, 'escuchar'),
    (13, 'parentesco'),
    (14, 'noviazgo'),
    (15, 'matrimonio'),
    (16, 'separación'),
    (17, 'alianza'),
    (18, 'comunidad'),
    (19, 'vecindad'),
    (20, 'cooperación'),
    (21, 'afectivo'),
    (22, 'aprecio'),
    (23, 'apego'),
    (24, 'distanciamiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Отношения, характер, психология / Конфликты, извинения, границы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Конфликты, извинения, границы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Конфликты, извинения, границы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'reproche'),
    (1, 'culpa'),
    (2, 'ceder'),
    (3, 'explicar'),
    (4, 'alejarse'),
    (5, 'consentir'),
    (6, 'tolerar'),
    (7, 'pactar'),
    (8, 'confrontación'),
    (9, 'distanciarse'),
    (10, 'ofender'),
    (11, 'resentimiento'),
    (12, 'invasión'),
    (13, 'reconciliarse'),
    (14, 'vulnerar'),
    (15, 'imponer'),
    (16, 'interrumpir'),
    (17, 'quejarse'),
    (18, 'reprochar'),
    (19, 'zanjar'),
    (20, 'desahogarse'),
    (21, 'desautorizar'),
    (22, 'incomodar'),
    (23, 'apaciguar'),
    (24, 'pactación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Отношения, характер, психология / Привычки, мотивация, стресс
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Привычки, мотивация, стресс'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Отношения, характер, психология'
    AND ws.title = 'Привычки, мотивация, стресс'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'estrés'),
    (1, 'recompensa'),
    (2, 'relajación'),
    (3, 'motivación'),
    (4, 'descanso'),
    (5, 'energía'),
    (6, 'presión'),
    (7, 'cambio'),
    (8, 'equilibrio'),
    (9, 'organizarse'),
    (10, 'concentrarse'),
    (11, 'recuperarse'),
    (12, 'priorizar'),
    (13, 'postergar'),
    (14, 'agotarse'),
    (15, 'autocontrol'),
    (16, 'perseverancia'),
    (17, 'agotamiento'),
    (18, 'sobrecarga'),
    (19, 'estímulo'),
    (20, 'voluntad'),
    (21, 'autocuidado'),
    (22, 'relajarse'),
    (23, 'desmotivación'),
    (24, 'resistencia'),
    (25, 'regularidad'),
    (26, 'procrastinar'),
    (27, 'desconexión'),
    (28, 'sobrellevar'),
    (29, 'respiración')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Медиа, культура, мнения / Новости и общественные темы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Новости и общественные темы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Новости и общественные темы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'actualidad'),
    (1, 'sociedad'),
    (2, 'política'),
    (3, 'gobierno'),
    (4, 'ley'),
    (5, 'ciudadano'),
    (6, 'protesta'),
    (7, 'debate'),
    (8, 'encuesta'),
    (9, 'crisis'),
    (10, 'reforma'),
    (11, 'prensa'),
    (12, 'reportaje'),
    (13, 'redacción'),
    (14, 'portada'),
    (15, 'editorial'),
    (16, 'corresponsal'),
    (17, 'entrevistar'),
    (18, 'audiencia'),
    (19, 'suceso'),
    (20, 'portavoz'),
    (21, 'manifestación'),
    (22, 'campaña'),
    (23, 'crónica'),
    (24, 'rueda')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Медиа, культура, мнения / Кино, сериалы, книги
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Кино, сериалы, книги'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Кино, сериалы, книги'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'personaje'),
    (1, 'guion'),
    (2, 'género'),
    (3, 'estreno'),
    (4, 'narrador'),
    (5, 'protagonista'),
    (6, 'recomendar'),
    (7, 'reseñar'),
    (8, 'adaptar'),
    (9, 'rodar'),
    (10, 'reparto'),
    (11, 'temporada'),
    (12, 'episodio'),
    (13, 'largometraje'),
    (14, 'cortometraje'),
    (15, 'documental'),
    (16, 'comedia'),
    (17, 'drama'),
    (18, 'suspense'),
    (19, 'doblaje'),
    (20, 'subtítulo'),
    (21, 'banda'),
    (22, 'tráiler'),
    (23, 'autora'),
    (24, 'montaje')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Медиа, культура, мнения / Музыка, искусство, события
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Музыка, искусство, события'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Музыка, искусство, события'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'pintura'),
    (1, 'escultura'),
    (2, 'público'),
    (3, 'espectáculo'),
    (4, 'instrumento'),
    (5, 'melodía'),
    (6, 'ritmo'),
    (7, 'coro'),
    (8, 'actuar'),
    (9, 'exponer'),
    (10, 'estrenar'),
    (11, 'lienzo'),
    (12, 'orquesta'),
    (13, 'recital'),
    (14, 'acústico'),
    (15, 'compositor'),
    (16, 'bailarina'),
    (17, 'inauguración'),
    (18, 'mural'),
    (19, 'partitura')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Медиа, культура, мнения / Отзывы, рецензии, впечатления
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Отзывы, рецензии, впечатления'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Отзывы, рецензии, впечатления'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'impresión'),
    (1, 'valoración'),
    (2, 'calidad'),
    (3, 'estilo'),
    (4, 'ambiente'),
    (5, 'puntuación'),
    (6, 'elogio'),
    (7, 'decepcionar'),
    (8, 'sorprender'),
    (9, 'valorar'),
    (10, 'memorable'),
    (11, 'agradable'),
    (12, 'flojo'),
    (13, 'excelente'),
    (14, 'mediocre'),
    (15, 'convincente'),
    (16, 'entretenido'),
    (17, 'emocionante'),
    (18, 'logrado'),
    (19, 'irregular'),
    (20, 'previsible'),
    (21, 'original'),
    (22, 'impactante'),
    (23, 'decepcionante'),
    (24, 'recomendable')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Медиа, культура, мнения / Аргументация вкуса
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Аргументация вкуса'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Медиа, культура, мнения'
    AND ws.title = 'Аргументация вкуса'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'trama'),
    (1, 'voz'),
    (2, 'imagen'),
    (3, 'tono'),
    (4, 'textura'),
    (5, 'estética'),
    (6, 'criterio'),
    (7, 'convencer'),
    (8, 'apreciar'),
    (9, 'subjetivo'),
    (10, 'objetivamente'),
    (11, 'coherente'),
    (12, 'atractivo'),
    (13, 'expresivo'),
    (14, 'intenso'),
    (15, 'sutil'),
    (16, 'exagerado'),
    (17, 'repetitivo'),
    (18, 'auténtico'),
    (19, 'artificial'),
    (20, 'profundidad'),
    (21, 'superficialidad'),
    (22, 'armonía'),
    (23, 'preferible'),
    (24, 'disfrutable')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Деньги, покупки, потребление / Доходы, расходы, бюджет
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Доходы, расходы, бюджет'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Доходы, расходы, бюджет'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'inversión'),
    (1, 'ahorro'),
    (2, 'préstamo'),
    (3, 'riqueza'),
    (4, 'pobreza'),
    (5, 'ahorrar'),
    (6, 'gastar'),
    (7, 'calcular'),
    (8, 'endeudarse'),
    (9, 'invertir'),
    (10, 'financiar'),
    (11, 'liquidez'),
    (12, 'nómina'),
    (13, 'neto'),
    (14, 'bruto'),
    (15, 'imprevisto'),
    (16, 'recorte'),
    (17, 'doméstica'),
    (18, 'patrimonio'),
    (19, 'desembolso'),
    (20, 'familiar'),
    (21, 'déficit'),
    (22, 'doméstico'),
    (23, 'excedente'),
    (24, 'económico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Деньги, покупки, потребление / Покупки, качество, возврат
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Покупки, качество, возврат'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Покупки, качество, возврат'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'producto'),
    (1, 'devolución'),
    (2, 'vendedor'),
    (3, 'comprador'),
    (4, 'etiqueta'),
    (5, 'material'),
    (6, 'marca'),
    (7, 'reclamar'),
    (8, 'duradero'),
    (9, 'defectuoso'),
    (10, 'embalaje'),
    (11, 'reembolso'),
    (12, 'sustitución'),
    (13, 'comprobante'),
    (14, 'legal'),
    (15, 'acabado'),
    (16, 'incorrecta'),
    (17, 'descosido'),
    (18, 'rayado'),
    (19, 'manchado'),
    (20, 'roto'),
    (21, 'caducado'),
    (22, 'comercial'),
    (23, 'parcial'),
    (24, 'gratuito')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Деньги, покупки, потребление / Банк, карты, переводы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Банк, карты, переводы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Банк, карты, переводы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'comisión'),
    (1, 'clave'),
    (2, 'operación'),
    (3, 'retirada'),
    (4, 'bancaria'),
    (5, 'cargo'),
    (6, 'validar'),
    (7, 'autorizar'),
    (8, 'banca'),
    (9, 'pin'),
    (10, 'débito'),
    (11, 'crédito'),
    (12, 'domiciliación'),
    (13, 'reintegro'),
    (14, 'beneficiario'),
    (15, 'bizum'),
    (16, 'domiciliado'),
    (17, 'secreta'),
    (18, 'móvil'),
    (19, 'automático')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Деньги, покупки, потребление / Рынок, цены, скидки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Рынок, цены, скидки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Рынок, цены, скидки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'rebaja'),
    (1, 'coste'),
    (2, 'valor'),
    (3, 'demanda'),
    (4, 'margen'),
    (5, 'promoción'),
    (6, 'escaparate'),
    (7, 'catálogo'),
    (8, 'encarecer'),
    (9, 'abaratar'),
    (10, 'ofertar'),
    (11, 'remate'),
    (12, 'cupón'),
    (13, 'mayorista'),
    (14, 'minorista'),
    (15, 'regateo'),
    (16, 'comparador'),
    (17, 'sobreprecio'),
    (18, 'puja'),
    (19, 'rebajado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Деньги, покупки, потребление / Потребительские проблемы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Потребительские проблемы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Деньги, покупки, потребление'
    AND ws.title = 'Потребительские проблемы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'engaño'),
    (1, 'soporte'),
    (2, 'compensación'),
    (3, 'compensar'),
    (4, 'reemplazar'),
    (5, 'reembolsar'),
    (6, 'estafar'),
    (7, 'incumplimiento'),
    (8, 'vencida'),
    (9, 'publicidad'),
    (10, 'engañosa'),
    (11, 'reclamaciones'),
    (12, 'abierta'),
    (13, 'dañado'),
    (14, 'cobro'),
    (15, 'indebido'),
    (16, 'erróneo'),
    (17, 'injustificado'),
    (18, 'trato'),
    (19, 'deficiente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Планирование поездки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Планирование поездки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Планирование поездки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'itinerario'),
    (1, 'transporte'),
    (2, 'excursión'),
    (3, 'escala'),
    (4, 'planificar'),
    (5, 'visitar'),
    (6, 'alojarse'),
    (7, 'recorrer'),
    (8, 'trayecto'),
    (9, 'agencia'),
    (10, 'folleto'),
    (11, 'previsto'),
    (12, 'ligero'),
    (13, 'guiada'),
    (14, 'pernoctar'),
    (15, 'reservarse'),
    (16, 'anticipar'),
    (17, 'presupuestar'),
    (18, 'empaquetar'),
    (19, 'escapada')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Транспортные проблемы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Транспортные проблемы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Транспортные проблемы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'atasco'),
    (1, 'multa'),
    (2, 'huelga'),
    (3, 'embarcar'),
    (4, 'retrasarse'),
    (5, 'atascarse'),
    (6, 'desviar'),
    (7, 'ferroviaria'),
    (8, 'overbooking'),
    (9, 'cancelarse'),
    (10, 'desviarse'),
    (11, 'demorado'),
    (12, 'averiarse'),
    (13, 'embotellamiento'),
    (14, 'colapso'),
    (15, 'retrasado'),
    (16, 'cancelado'),
    (17, 'transbordar'),
    (18, 'reprogramar'),
    (19, 'ferroviario')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Районы и инфраструктура
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Районы и инфраструктура'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Районы и инфраструктура'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'infraestructura'),
    (1, 'comercio'),
    (2, 'accesible'),
    (3, 'céntrico'),
    (4, 'residencial'),
    (5, 'peatonal'),
    (6, 'periférico'),
    (7, 'cercano'),
    (8, 'urbanizar'),
    (9, 'renovar'),
    (10, 'distrito'),
    (11, 'equipamiento'),
    (12, 'alumbrado'),
    (13, 'carril'),
    (14, 'alcantarillado'),
    (15, 'acerado'),
    (16, 'carrilbici'),
    (17, 'urbanización'),
    (18, 'arbolado'),
    (19, 'señalización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Природа, экология, погода B1
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Природа, экология, погода B1'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Природа, экология, погода B1'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ecología'),
    (1, 'contaminación'),
    (2, 'reciclaje'),
    (3, 'residuo'),
    (4, 'sequía'),
    (5, 'sostenibilidad'),
    (6, 'emisión'),
    (7, 'conservar'),
    (8, 'reciclar'),
    (9, 'contaminar'),
    (10, 'paisaje'),
    (11, 'biodiversidad'),
    (12, 'vertido'),
    (13, 'humareda'),
    (14, 'erosión'),
    (15, 'reforestación'),
    (16, 'biodegradable'),
    (17, 'ecológico'),
    (18, 'atmosférico'),
    (19, 'nuboso')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Жильё, соседи, городская жизнь
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Жильё, соседи, городская жизнь'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Жильё, соседи, городская жизнь'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'vivienda'),
    (1, 'edificio'),
    (2, 'mudanza'),
    (3, 'escalera'),
    (4, 'patio'),
    (5, 'reformar'),
    (6, 'convivir'),
    (7, 'derrama'),
    (8, 'vecinal'),
    (9, 'urbana'),
    (10, 'vecindario'),
    (11, 'casero'),
    (12, 'copropietario'),
    (13, 'portería'),
    (14, 'comunitario'),
    (15, 'rellano'),
    (16, 'humedades'),
    (17, 'domiciliario'),
    (18, 'azotea'),
    (19, 'fachada')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Путешествия, город, среда / Впечатления от мест
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Впечатления от мест'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Путешествия, город, среда'
    AND ws.title = 'Впечатления от мест'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'belleza'),
    (1, 'tranquilidad'),
    (2, 'encanto'),
    (3, 'vista'),
    (4, 'descubrir'),
    (5, 'disfrutar'),
    (6, 'admirar'),
    (7, 'panorámica'),
    (8, 'mirador'),
    (9, 'pintoresco'),
    (10, 'acogedor'),
    (11, 'vibrante'),
    (12, 'turístico'),
    (13, 'inolvidable'),
    (14, 'apacible'),
    (15, 'bullicioso'),
    (16, 'descuidado'),
    (17, 'animado'),
    (18, 'silencioso'),
    (19, 'monumental')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Устройства и настройки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Устройства и настройки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Устройства и настройки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'dispositivo'),
    (1, 'configuración'),
    (2, 'batería'),
    (3, 'cable'),
    (4, 'cargador'),
    (5, 'ajuste'),
    (6, 'sistema'),
    (7, 'adaptador'),
    (8, 'cargarse'),
    (9, 'enchufar'),
    (10, 'emparejar'),
    (11, 'bluetooth'),
    (12, 'inalámbrico'),
    (13, 'táctil'),
    (14, 'almacenamiento'),
    (15, 'brillo'),
    (16, 'volumen'),
    (17, 'auricular'),
    (18, 'cargable'),
    (19, 'calibrar'),
    (20, 'inteligente'),
    (21, 'sensor'),
    (22, 'funda'),
    (23, 'protector'),
    (24, 'router')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Приложения, аккаунты, безопасность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Приложения, аккаунты, безопасность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Приложения, аккаунты, безопасность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'verificación'),
    (1, 'autenticación'),
    (2, 'configurar'),
    (3, 'iniciar'),
    (4, 'restablecer'),
    (5, 'cifrado'),
    (6, 'autenticador'),
    (7, 'biometría'),
    (8, 'desbloqueo'),
    (9, 'verificador'),
    (10, 'código'),
    (11, 'doble'),
    (12, 'factor'),
    (13, 'antiphishing'),
    (14, 'limitado'),
    (15, 'token'),
    (16, 'digital'),
    (17, 'permisos'),
    (18, 'identidad'),
    (19, 'caducar'),
    (20, 'anonimato'),
    (21, 'suplantación'),
    (22, 'intruso'),
    (23, 'sospechoso'),
    (24, 'desbloquear')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Разработка и IT-работа: базовый слой
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Разработка и IT-работа: базовый слой'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Разработка и IT-работа: базовый слой'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'desarrollo'),
    (1, 'base'),
    (2, 'repositorio'),
    (3, 'rama'),
    (4, 'entorno'),
    (5, 'programar'),
    (6, 'diseñar'),
    (7, 'desplegar'),
    (8, 'compilar'),
    (9, 'depurar'),
    (10, 'integrar'),
    (11, 'backend'),
    (12, 'frontend'),
    (13, 'script'),
    (14, 'framework'),
    (15, 'librería'),
    (16, 'commit'),
    (17, 'bug'),
    (18, 'técnica'),
    (19, 'despliegue'),
    (20, 'local'),
    (21, 'virtual'),
    (22, 'integración'),
    (23, 'continua'),
    (24, 'api'),
    (25, 'endpoint'),
    (26, 'maqueta'),
    (27, 'prototipo'),
    (28, 'componente'),
    (29, 'compilación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Данные, файлы, сервисы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Данные, файлы, сервисы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Данные, файлы, сервисы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'descarga'),
    (1, 'carga'),
    (2, 'formato'),
    (3, 'respaldo'),
    (4, 'exportación'),
    (5, 'directorio'),
    (6, 'comprimir'),
    (7, 'descomprimir'),
    (8, 'migrar'),
    (9, 'restaurar'),
    (10, 'metadato'),
    (11, 'binario'),
    (12, 'plantilla'),
    (13, 'sincronización'),
    (14, 'remota'),
    (15, 'automática'),
    (16, 'manual'),
    (17, 'externo'),
    (18, 'restauración'),
    (19, 'dataset'),
    (20, 'sincronizado'),
    (21, 'exportable'),
    (22, 'importable'),
    (23, 'duplicado'),
    (24, 'incremental')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Ошибки, поддержка, инструкции
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Ошибки, поддержка, инструкции'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Ошибки, поддержка, инструкции'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'instrucción'),
    (1, 'paso'),
    (2, 'reinicio'),
    (3, 'reportar'),
    (4, 'error'),
    (5, 'captura'),
    (6, 'tutorial'),
    (7, 'procedimiento'),
    (8, 'parche'),
    (9, 'contactar'),
    (10, 'escalar'),
    (11, 'solucionador'),
    (12, 'congelada'),
    (13, 'forzado'),
    (14, 'modo'),
    (15, 'ticket'),
    (16, 'asistencia'),
    (17, 'depuración'),
    (18, 'pantallazo'),
    (19, 'solucionable'),
    (20, 'reiniciable'),
    (21, 'guiado'),
    (22, 'alerta'),
    (23, 'documentar'),
    (24, 'reintentar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B1 / Технологии и интернет / Онлайн-коммуникация и контент
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Онлайн-коммуникация и контент'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B1'
    AND cat.name = 'Технологии и интернет'
    AND ws.title = 'Онлайн-коммуникация и контент'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'contenido'),
    (1, 'directo'),
    (2, 'difusión'),
    (3, 'moderación'),
    (4, 'privado'),
    (5, 'hilo'),
    (6, 'viral'),
    (7, 'visualización'),
    (8, 'reacción'),
    (9, 'retransmitir'),
    (10, 'grabación'),
    (11, 'retransmisión'),
    (12, 'vivo'),
    (13, 'moderador'),
    (14, 'creador'),
    (15, 'programada'),
    (16, 'fijado'),
    (17, 'métrica'),
    (18, 'interacción'),
    (19, 'post'),
    (20, 'multimedia'),
    (21, 'streaming'),
    (22, 'moderable'),
    (23, 'transmisor'),
    (24, 'suscriptor'),
    (25, 'notificador'),
    (26, 'creadora'),
    (27, 'retransmisor'),
    (28, 'reaccionar'),
    (29, 'difundir')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Аргументация и логические связки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Аргументация и логические связки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Аргументация и логические связки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'argumentación'),
    (1, 'premisa'),
    (2, 'razonamiento'),
    (3, 'inferencia'),
    (4, 'inducción'),
    (5, 'contraargumento'),
    (6, 'réplica'),
    (7, 'coherencia'),
    (8, 'cohesión'),
    (9, 'postura'),
    (10, 'fundamento'),
    (11, 'perspectiva'),
    (12, 'refutar'),
    (13, 'sostener'),
    (14, 'demostrar'),
    (15, 'conclusión'),
    (16, 'lógica'),
    (17, 'central'),
    (18, 'razonable'),
    (19, 'debatible'),
    (20, 'sólido'),
    (21, 'consistente'),
    (22, 'inconsistente'),
    (23, 'refutación'),
    (24, 'demostración'),
    (25, 'inferir'),
    (26, 'deducir'),
    (27, 'inducir'),
    (28, 'fundamentar'),
    (29, 'articulación'),
    (30, 'conector'),
    (31, 'refutable'),
    (32, 'ilación'),
    (33, 'silogismo'),
    (34, 'proposición'),
    (35, 'aseveración'),
    (36, 'defendible'),
    (37, 'argumentativo'),
    (38, 'contraejemplo'),
    (39, 'lógico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Вероятность, обязанность, допущение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Вероятность, обязанность, допущение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Вероятность, обязанность, допущение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'obligación'),
    (1, 'necesidad'),
    (2, 'prohibición'),
    (3, 'eventualidad'),
    (4, 'factible'),
    (5, 'viable'),
    (6, 'obligatorio'),
    (7, 'opcional'),
    (8, 'probable'),
    (9, 'admisible'),
    (10, 'ineludible'),
    (11, 'imprescindible'),
    (12, 'conveniente'),
    (13, 'contingente'),
    (14, 'exigible'),
    (15, 'permitido'),
    (16, 'prohibido'),
    (17, 'suponer'),
    (18, 'prever'),
    (19, 'estimar'),
    (20, 'conceder'),
    (21, 'descartar'),
    (22, 'obligatoriedad'),
    (23, 'presunción'),
    (24, 'admisibilidad'),
    (25, 'viabilidad'),
    (26, 'factibilidad'),
    (27, 'permisividad'),
    (28, 'imposición'),
    (29, 'imperativo'),
    (30, 'previo'),
    (31, 'moral'),
    (32, 'previsibilidad'),
    (33, 'riesgo'),
    (34, 'asumible')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Причинно-следственные цепочки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Причинно-следственные цепочки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Причинно-следственные цепочки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'origen'),
    (1, 'impacto'),
    (2, 'cadena'),
    (3, 'proceso'),
    (4, 'variable'),
    (5, 'correlación'),
    (6, 'desencadenar'),
    (7, 'influir'),
    (8, 'repercutir'),
    (9, 'depender'),
    (10, 'incidir'),
    (11, 'causar'),
    (12, 'propiciar'),
    (13, 'producir'),
    (14, 'multiplicar'),
    (15, 'agravar'),
    (16, 'atenuar'),
    (17, 'concatenación'),
    (18, 'causal'),
    (19, 'nexo'),
    (20, 'indirecta'),
    (21, 'ramificación'),
    (22, 'retroalimentación'),
    (23, 'secuela'),
    (24, 'secundaria'),
    (25, 'dominó'),
    (26, 'acumulación'),
    (27, 'condicionamiento'),
    (28, 'catalizador'),
    (29, 'detonación'),
    (30, 'propagación'),
    (31, 'ramificarse'),
    (32, 'intensificar'),
    (33, 'moderar'),
    (34, 'desencadenamiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Контраргументы и нюансирование
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Контраргументы и нюансирование'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Контраргументы и нюансирование'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ambigüedad'),
    (1, 'complejidad'),
    (2, 'paradoja'),
    (3, 'contradicción'),
    (4, 'cuestionar'),
    (5, 'suavizar'),
    (6, 'objetar'),
    (7, 'contraponer'),
    (8, 'problematizar'),
    (9, 'refinar'),
    (10, 'complejizar'),
    (11, 'contextualizar'),
    (12, 'restringir'),
    (13, 'atenuante'),
    (14, 'reparo'),
    (15, 'cautela'),
    (16, 'metodológica'),
    (17, 'terminológica'),
    (18, 'contralectura'),
    (19, 'alternativo'),
    (20, 'sesgo'),
    (21, 'simplificación'),
    (22, 'reduccionismo'),
    (23, 'sobregeneralización'),
    (24, 'ambivalencia'),
    (25, 'concesivo'),
    (26, 'problemático'),
    (27, 'discutible'),
    (28, 'objecionable'),
    (29, 'rebatible'),
    (30, 'discutibilidad'),
    (31, 'cauteloso'),
    (32, 'condicionado'),
    (33, 'matizable'),
    (34, 'revisable')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Дискуссия, дебаты, переговоры
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Дискуссия, дебаты, переговоры'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Дискуссия, дебаты, переговоры'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'negociación'),
    (1, 'consenso'),
    (2, 'persuasión'),
    (3, 'pacto'),
    (4, 'intervención'),
    (5, 'interlocutor'),
    (6, 'deliberación'),
    (7, 'asamblea'),
    (8, 'foro'),
    (9, 'persuadir'),
    (10, 'intervenir'),
    (11, 'discrepar'),
    (12, 'transigir'),
    (13, 'conciliar'),
    (14, 'interlocución'),
    (15, 'deliberar'),
    (16, 'deliberativo'),
    (17, 'negociador'),
    (18, 'mediador'),
    (19, 'conciliación'),
    (20, 'arbitraje'),
    (21, 'contrapropuesta'),
    (22, 'moción'),
    (23, 'palabra'),
    (24, 'común'),
    (25, 'punto'),
    (26, 'muerto'),
    (27, 'consensuar'),
    (28, 'disentir'),
    (29, 'argumentar'),
    (30, 'rebatir'),
    (31, 'polemizar'),
    (32, 'pública'),
    (33, 'conciliador'),
    (34, 'negociable'),
    (35, 'regatear'),
    (36, 'deliberante'),
    (37, 'discrepante'),
    (38, 'negociadora'),
    (39, 'contraparte')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Формальный и нейтральный регистр
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Формальный и нейтральный регистр'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Формальный и нейтральный регистр'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'formalidad'),
    (1, 'neutralidad'),
    (2, 'cortesía'),
    (3, 'declaración'),
    (4, 'protocolario'),
    (5, 'institucional'),
    (6, 'neutral'),
    (7, 'respetuoso'),
    (8, 'manifestar'),
    (9, 'declarar'),
    (10, 'formulación'),
    (11, 'impersonal'),
    (12, 'pasiva'),
    (13, 'nominalización'),
    (14, 'atenuador'),
    (15, 'deferencia'),
    (16, 'encabezamiento'),
    (17, 'atentamente'),
    (18, 'cordialmente'),
    (19, 'compareciente'),
    (20, 'suscrito'),
    (21, 'interesado'),
    (22, 'presente'),
    (23, 'procedente'),
    (24, 'pertinente'),
    (25, 'conforme'),
    (26, 'preceptivo'),
    (27, 'reglado'),
    (28, 'normalizado'),
    (29, 'acreditar'),
    (30, 'cumplimentar'),
    (31, 'tramitar'),
    (32, 'elevar'),
    (33, 'diligenciado'),
    (34, 'protocolizado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Связность, аргументация, модальность / Резюмирование и выводы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Резюмирование и выводы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Связность, аргументация, модальность'
    AND ws.title = 'Резюмирование и выводы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'síntesis'),
    (1, 'recapitulación'),
    (2, 'panorama'),
    (3, 'hallazgo'),
    (4, 'implicación'),
    (5, 'enseñanza'),
    (6, 'moraleja'),
    (7, 'global'),
    (8, 'provisional'),
    (9, 'concluir'),
    (10, 'recapitular'),
    (11, 'extraer'),
    (12, 'proyectar'),
    (13, 'recapitulativo'),
    (14, 'conclusivo'),
    (15, 'argumental'),
    (16, 'corolario'),
    (17, 'principal'),
    (18, 'derivada'),
    (19, 'general'),
    (20, 'extraíble'),
    (21, 'sintetizador'),
    (22, 'compendio'),
    (23, 'epílogo'),
    (24, 'recapitulador'),
    (25, 'finalizador'),
    (26, 'desenlace'),
    (27, 'discursivo'),
    (28, 'condensación'),
    (29, 'clausura')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Качество, эффективность, результат
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Качество, эффективность, результат'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Качество, эффективность, результат'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'eficiencia'),
    (1, 'eficacia'),
    (2, 'utilidad'),
    (3, 'alcance'),
    (4, 'mejora'),
    (5, 'estándar'),
    (6, 'optimización'),
    (7, 'excelencia'),
    (8, 'ventaja'),
    (9, 'desventaja'),
    (10, 'fiable'),
    (11, 'optimizar'),
    (12, 'productivo'),
    (13, 'eficiente'),
    (14, 'aprovechamiento'),
    (15, 'satisfactorio'),
    (16, 'satisfactoriedad'),
    (17, 'idoneidad'),
    (18, 'robustez'),
    (19, 'fiabilidad'),
    (20, 'funcionalidad'),
    (21, 'consistencia'),
    (22, 'operacional'),
    (23, 'óptimo'),
    (24, 'medible'),
    (25, 'real'),
    (26, 'tangible'),
    (27, 'optimizable'),
    (28, 'rentable'),
    (29, 'sostenible'),
    (30, 'robusto'),
    (31, 'funcional'),
    (32, 'operatividad'),
    (33, 'comprobable'),
    (34, 'idóneo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Изменение, развитие, тенденции
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Изменение, развитие, тенденции'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Изменение, развитие, тенденции'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'tendencia'),
    (1, 'transformación'),
    (2, 'transición'),
    (3, 'crecimiento'),
    (4, 'declive'),
    (5, 'retroceso'),
    (6, 'innovación'),
    (7, 'adaptación'),
    (8, 'modernización'),
    (9, 'acelerar'),
    (10, 'ralentizar'),
    (11, 'evolucionar'),
    (12, 'transformar'),
    (13, 'consolidar'),
    (14, 'mutación'),
    (15, 'giro'),
    (16, 'renovación'),
    (17, 'expansión'),
    (18, 'contracción'),
    (19, 'estancamiento'),
    (20, 'deterioro'),
    (21, 'progresión'),
    (22, 'regresión'),
    (23, 'reconfiguración'),
    (24, 'reconversión'),
    (25, 'desplazamiento'),
    (26, 'auge'),
    (27, 'desaceleración'),
    (28, 'aceleración'),
    (29, 'fluctuación'),
    (30, 'variación'),
    (31, 'viraje'),
    (32, 'maduración'),
    (33, 'consolidación'),
    (34, 'gradualidad')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Проблема, риск, решение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Проблема, риск, решение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Проблема, риск, решение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'desafío'),
    (1, 'obstáculo'),
    (2, 'mitigación'),
    (3, 'vulnerabilidad'),
    (4, 'reducir'),
    (5, 'afrontar'),
    (6, 'detectar'),
    (7, 'contingencia'),
    (8, 'crítico'),
    (9, 'abordaje'),
    (10, 'dificultad'),
    (11, 'latente'),
    (12, 'grave'),
    (13, 'resolución'),
    (14, 'mitigable'),
    (15, 'neutralizar'),
    (16, 'minimizar'),
    (17, 'contener'),
    (18, 'residual'),
    (19, 'preventiva'),
    (20, 'escollo'),
    (21, 'contratiempo'),
    (22, 'fragilidad'),
    (23, 'emergente'),
    (24, 'operativa'),
    (25, 'subsanable'),
    (26, 'reparable'),
    (27, 'reversible'),
    (28, 'irreversible'),
    (29, 'paliar'),
    (30, 'remediable'),
    (31, 'evitable'),
    (32, 'prevenible'),
    (33, 'desactivable'),
    (34, 'saneable')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Ценности, мораль, ответственность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Ценности, мораль, ответственность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Ценности, мораль, ответственность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ética'),
    (1, 'justicia'),
    (2, 'igualdad'),
    (3, 'libertad'),
    (4, 'dignidad'),
    (5, 'solidaridad'),
    (6, 'honestidad'),
    (7, 'transparencia'),
    (8, 'mérito'),
    (9, 'conciencia'),
    (10, 'respetar'),
    (11, 'integridad'),
    (12, 'equidad'),
    (13, 'virtud'),
    (14, 'social'),
    (15, 'cívico'),
    (16, 'ético'),
    (17, 'rectitud'),
    (18, 'bien'),
    (19, 'cívica'),
    (20, 'corresponsabilidad'),
    (21, 'intelectual'),
    (22, 'colectiva'),
    (23, 'personal'),
    (24, 'legitimidad'),
    (25, 'moralidad'),
    (26, 'deontología'),
    (27, 'altruismo'),
    (28, 'civismo'),
    (29, 'imparcialidad'),
    (30, 'probidad'),
    (31, 'honradez'),
    (32, 'decencia'),
    (33, 'humana'),
    (34, 'distributiva')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Восприятие, интерпретация, отношение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Восприятие, интерпретация, отношение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Восприятие, интерпретация, отношение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'percepción'),
    (1, 'interpretación'),
    (2, 'actitud'),
    (3, 'mirada'),
    (4, 'sentido'),
    (5, 'significado'),
    (6, 'percibir'),
    (7, 'subjetividad'),
    (8, 'observador'),
    (9, 'receptividad'),
    (10, 'sensibilidad'),
    (11, 'apreciación'),
    (12, 'subjetiva'),
    (13, 'predisposición'),
    (14, 'receptivo'),
    (15, 'interpretativo'),
    (16, 'simbólico'),
    (17, 'connotación'),
    (18, 'emocional'),
    (19, 'cognitivo'),
    (20, 'observación'),
    (21, 'apreciativa'),
    (22, 'externa'),
    (23, 'sensorial'),
    (24, 'simbólica'),
    (25, 'marco'),
    (26, 'subjetivismo'),
    (27, 'observacional'),
    (28, 'disposición'),
    (29, 'mental'),
    (30, 'hermenéutica'),
    (31, 'interpretabilidad'),
    (32, 'predisponer'),
    (33, 'percibido'),
    (34, 'percibible')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Сравнение, приоритеты, компромиссы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Сравнение, приоритеты, компромиссы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Сравнение, приоритеты, компромиссы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'comparación'),
    (1, 'renuncia'),
    (2, 'jerarquía'),
    (3, 'sacrificio'),
    (4, 'opción'),
    (5, 'seleccionar'),
    (6, 'equilibrar'),
    (7, 'ponderar'),
    (8, 'escoger'),
    (9, 'secundario'),
    (10, 'esencial'),
    (11, 'relativa'),
    (12, 'jerarquización'),
    (13, 'oportunidad'),
    (14, 'mutua'),
    (15, 'prelación'),
    (16, 'decisivo'),
    (17, 'estratégica'),
    (18, 'elección'),
    (19, 'prioritaria'),
    (20, 'ponderación'),
    (21, 'comparativa'),
    (22, 'priorización'),
    (23, 'jerarquizable'),
    (24, 'relativo'),
    (25, 'inestable'),
    (26, 'compensatorio'),
    (27, 'renunciable'),
    (28, 'prioritario'),
    (29, 'prescindible'),
    (30, 'preferente'),
    (31, 'subsidiario'),
    (32, 'intercambiable'),
    (33, 'compensable'),
    (34, 'subóptimo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Абстрактные понятия и оценка / Общеупотребительные метафорические слова
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Общеупотребительные метафорические слова'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Абстрактные понятия и оценка'
    AND ws.title = 'Общеупотребительные метафорические слова'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'raíz'),
    (1, 'núcleo'),
    (2, 'motor'),
    (3, 'freno'),
    (4, 'impulso'),
    (5, 'barrera'),
    (6, 'horizonte'),
    (7, 'eje'),
    (8, 'peso'),
    (9, 'luz'),
    (10, 'sombra'),
    (11, 'rumbo'),
    (12, 'tejido'),
    (13, 'capa'),
    (14, 'palanca'),
    (15, 'ancla'),
    (16, 'brújula'),
    (17, 'grieta'),
    (18, 'umbral'),
    (19, 'suelo'),
    (20, 'columna'),
    (21, 'pilar'),
    (22, 'engranaje'),
    (23, 'bisagra'),
    (24, 'resorte'),
    (25, 'lastre'),
    (26, 'trampolín'),
    (27, 'telón'),
    (28, 'lente'),
    (29, 'interna'),
    (30, 'timón'),
    (31, 'nudo'),
    (32, 'cauce'),
    (33, 'corriente'),
    (34, 'pivote'),
    (35, 'cimiento'),
    (36, 'vertebral'),
    (37, 'telaraña'),
    (38, 'andamiaje'),
    (39, 'esqueleto')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Управление проектами
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Управление проектами'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Управление проектами'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'gestión'),
    (1, 'ejecución'),
    (2, 'desviación'),
    (3, 'iteración'),
    (4, 'delegar'),
    (5, 'secuenciar'),
    (6, 'calendarizar'),
    (7, 'controlar'),
    (8, 'monitorear'),
    (9, 'alinear'),
    (10, 'replanificar'),
    (11, 'dimensionar'),
    (12, 'priorizable'),
    (13, 'estimativo'),
    (14, 'entregabilidad'),
    (15, 'reasignar'),
    (16, 'patrocinador'),
    (17, 'gobernanza'),
    (18, 'subproyecto'),
    (19, 'trazabilidad'),
    (20, 'intermedio'),
    (21, 'transversal'),
    (22, 'dimensionamiento'),
    (23, 'secuenciación'),
    (24, 'calendarización'),
    (25, 'ágil'),
    (26, 'retrospectiva'),
    (27, 'refinamiento'),
    (28, 'estimable'),
    (29, 'comprometido'),
    (30, 'bloqueante'),
    (31, 'desviable'),
    (32, 'replanificación'),
    (33, 'patrocinio'),
    (34, 'reasignable'),
    (35, 'coordinación'),
    (36, 'ejecutiva'),
    (37, 'matricial'),
    (38, 'planificador'),
    (39, 'programador'),
    (40, 'iterativo'),
    (41, 'trazador'),
    (42, 'ejecutor'),
    (43, 'ejecutable'),
    (44, 'reasignación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Команда и лидерство
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Команда и лидерство'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Команда и лидерство'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'talento'),
    (1, 'mentor'),
    (2, 'reconocimiento'),
    (3, 'participación'),
    (4, 'dirigir'),
    (5, 'inspirar'),
    (6, 'empoderar'),
    (7, 'motivar'),
    (8, 'mentoría'),
    (9, 'horizontal'),
    (10, 'situacional'),
    (11, 'multidisciplinar'),
    (12, 'integrador'),
    (13, 'pertenencia'),
    (14, 'facilitador'),
    (15, 'cohesionador'),
    (16, 'delegante'),
    (17, 'motivacional'),
    (18, 'colaborativo'),
    (19, 'participativo'),
    (20, 'corresponsable'),
    (21, 'mentorizar'),
    (22, 'tutelar'),
    (23, 'supervisión'),
    (24, 'acompañamiento'),
    (25, 'empoderamiento'),
    (26, 'colectivo'),
    (27, 'escucha'),
    (28, 'activa'),
    (29, 'compartido'),
    (30, 'nuclear'),
    (31, 'ampliado'),
    (32, 'sinergia'),
    (33, 'alineamiento'),
    (34, 'cultural'),
    (35, 'dinamizador'),
    (36, 'cohesionante'),
    (37, 'efectiva'),
    (38, 'inclusivo'),
    (39, 'grupal')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Стратегия, цели, KPI
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Стратегия, цели, KPI'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Стратегия, цели, KPI'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'visión'),
    (1, 'misión'),
    (2, 'posicionamiento'),
    (3, 'alineación'),
    (4, 'táctico'),
    (5, 'táctica'),
    (6, 'programático'),
    (7, 'directriz'),
    (8, 'estratégico'),
    (9, 'competitiva'),
    (10, 'accionable'),
    (11, 'cuadro'),
    (12, 'mando'),
    (13, 'operativo'),
    (14, 'foco'),
    (15, 'cuantificable'),
    (16, 'corporativa'),
    (17, 'direccionamiento'),
    (18, 'competitivo'),
    (19, 'posicionador'),
    (20, 'metas'),
    (21, 'cuantificables'),
    (22, 'proyección'),
    (23, 'operativizar'),
    (24, 'instrumentalizar'),
    (25, 'desplegable'),
    (26, 'focalización'),
    (27, 'diferenciación'),
    (28, 'escalamiento'),
    (29, 'compartida'),
    (30, 'prospectiva'),
    (31, 'estratega'),
    (32, 'parametrizar'),
    (33, 'corporativo'),
    (34, 'direccionable'),
    (35, 'direccionalidad'),
    (36, 'organizacional'),
    (37, 'cuantificado'),
    (38, 'tabulación'),
    (39, 'operativas')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Продажи, клиенты, рынок
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Продажи, клиенты, рынок'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Продажи, клиенты, рынок'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'venta'),
    (1, 'segmento'),
    (2, 'fidelización'),
    (3, 'captación'),
    (4, 'conversión'),
    (5, 'posicionar'),
    (6, 'prospección'),
    (7, 'embudo'),
    (8, 'prospector'),
    (9, 'posventa'),
    (10, 'cartera'),
    (11, 'potencial'),
    (12, 'segmentación'),
    (13, 'canalización'),
    (14, 'fidelizar'),
    (15, 'captar'),
    (16, 'retener'),
    (17, 'monetizar'),
    (18, 'compraventa'),
    (19, 'preventa'),
    (20, 'comercialización'),
    (21, 'mercadeo'),
    (22, 'comisionista'),
    (23, 'clientela'),
    (24, 'nicho'),
    (25, 'publicitario'),
    (26, 'pujar'),
    (27, 'tarifar'),
    (28, 'revendedor'),
    (29, 'distribuidor'),
    (30, 'ofertante'),
    (31, 'demandante'),
    (32, 'monetización'),
    (33, 'comercializable'),
    (34, 'promocionar'),
    (35, 'bonificar'),
    (36, 'comerciar'),
    (37, 'comercializador'),
    (38, 'ofertador'),
    (39, 'captador')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Договорные и юридические рабочие слова
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Договорные и юридические рабочие слова'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Договорные и юридические рабочие слова'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'derecho'),
    (1, 'confidencialidad'),
    (2, 'litigio'),
    (3, 'parte'),
    (4, 'estipular'),
    (5, 'indemnizar'),
    (6, 'rescindir'),
    (7, 'prorrogar'),
    (8, 'anexo'),
    (9, 'penalización'),
    (10, 'exclusividad'),
    (11, 'jurisdicción'),
    (12, 'contractual'),
    (13, 'contractualidad'),
    (14, 'estipulación'),
    (15, 'indemnización'),
    (16, 'penalidad'),
    (17, 'rescindible'),
    (18, 'prorrogable'),
    (19, 'vinculante'),
    (20, 'vinculabilidad'),
    (21, 'cesionario'),
    (22, 'cedente'),
    (23, 'licenciatario'),
    (24, 'licenciante'),
    (25, 'subcontrata'),
    (26, 'subcontratista'),
    (27, 'adjudicatario'),
    (28, 'adjudicación'),
    (29, 'confidencial'),
    (30, 'territorial'),
    (31, 'licenciamiento'),
    (32, 'adenda'),
    (33, 'adendado'),
    (34, 'cláusulado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / HR, performance, оценка
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'HR, performance, оценка'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'HR, performance, оценка'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'rotación'),
    (1, 'incentivo'),
    (2, 'salarial'),
    (3, 'absentismo'),
    (4, 'cantera'),
    (5, 'despido'),
    (6, 'ascenso'),
    (7, 'laboral'),
    (8, 'profesional'),
    (9, 'periódica'),
    (10, 'voluntaria'),
    (11, 'bonificación'),
    (12, 'desvinculación'),
    (13, 'interno'),
    (14, 'constructivo'),
    (15, 'capacitación'),
    (16, 'individual'),
    (17, 'empleabilidad'),
    (18, 'evaluador'),
    (19, 'evaluable'),
    (20, 'evaluativo'),
    (21, 'retribución'),
    (22, 'bonificable'),
    (23, 'incentivable'),
    (24, 'relevo'),
    (25, 'competencias'),
    (26, 'actitudinal'),
    (27, 'desvincular'),
    (28, 'recolocación'),
    (29, 'evaluabilidad'),
    (30, 'empleable'),
    (31, 'promocionable'),
    (32, 'bonificador'),
    (33, 'cualitativa'),
    (34, 'cuantitativa')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Конфликты, переговоры, решения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Конфликты, переговоры, решения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Конфликты, переговоры, решения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'disputa'),
    (1, 'escalada'),
    (2, 'posición'),
    (3, 'arbitrar'),
    (4, 'confrontar'),
    (5, 'desescalar'),
    (6, 'desescalada'),
    (7, 'facilitación'),
    (8, 'transacción'),
    (9, 'persistente'),
    (10, 'dura'),
    (11, 'colaborativa'),
    (12, 'pactada'),
    (13, 'negociada'),
    (14, 'negociabilidad'),
    (15, 'conciliable'),
    (16, 'transaccional'),
    (17, 'arbitral'),
    (18, 'contendiente'),
    (19, 'enfrentamiento'),
    (20, 'impasse'),
    (21, 'mediable'),
    (22, 'transigencia'),
    (23, 'avenencia'),
    (24, 'avenible'),
    (25, 'litigioso'),
    (26, 'litigiosidad'),
    (27, 'contrapuesto'),
    (28, 'polarización'),
    (29, 'hostilidad'),
    (30, 'confrontativo'),
    (31, 'conciliatorio'),
    (32, 'impugnación'),
    (33, 'impugnar'),
    (34, 'contencioso')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Работа, бизнес, управление / Удалёнка и международная работа
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Удалёнка и международная работа'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Работа, бизнес, управление'
    AND ws.title = 'Удалёнка и международная работа'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'huso'),
    (1, 'videoconferencia'),
    (2, 'asincronía'),
    (3, 'internacional'),
    (4, 'flexible'),
    (5, 'distribuido'),
    (6, 'nómada'),
    (7, 'transfronterizo'),
    (8, 'multinacional'),
    (9, 'intercultural'),
    (10, 'remotamente'),
    (11, 'desfase'),
    (12, 'itinerante'),
    (13, 'híbrido'),
    (14, 'trabajo'),
    (15, 'externalización'),
    (16, 'internacionalización'),
    (17, 'teletrabajador'),
    (18, 'expatriación'),
    (19, 'deslocalización'),
    (20, 'telepresencia'),
    (21, 'teleconferencia'),
    (22, 'multinacionalidad'),
    (23, 'plurilingüe'),
    (24, 'bicultural'),
    (25, 'transnacional'),
    (26, 'externalizado'),
    (27, 'subcontratado'),
    (28, 'movilidad'),
    (29, 'transfronteriza')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Разработка ПО и архитектура
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Разработка ПО и архитектура'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Разработка ПО и архитектура'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'arquitectura'),
    (1, 'módulo'),
    (2, 'escalabilidad'),
    (3, 'contenedor'),
    (4, 'automatización'),
    (5, 'usabilidad'),
    (6, 'digitalización'),
    (7, 'microservicio'),
    (8, 'implementación'),
    (9, 'mantenibilidad'),
    (10, 'desacoplar'),
    (11, 'encapsular'),
    (12, 'modularizar'),
    (13, 'versionar'),
    (14, 'refactorizar'),
    (15, 'orquestar'),
    (16, 'interoperable'),
    (17, 'acoplamiento'),
    (18, 'monolito'),
    (19, 'modularidad'),
    (20, 'abstracción'),
    (21, 'patrón'),
    (22, 'arquitectónico'),
    (23, 'escalable'),
    (24, 'extensible'),
    (25, 'testeable'),
    (26, 'mantenible'),
    (27, 'desacoplado'),
    (28, 'modularizado'),
    (29, 'refactorización'),
    (30, 'encapsulación'),
    (31, 'interfazado'),
    (32, 'versionado'),
    (33, 'empaquetado'),
    (34, 'interoperabilidad'),
    (35, 'extensibilidad'),
    (36, 'portabilidad'),
    (37, 'integrabilidad'),
    (38, 'arquitecturable'),
    (39, 'componibilidad'),
    (40, 'modularizable'),
    (41, 'portable'),
    (42, 'integrable'),
    (43, 'versionable'),
    (44, 'testeabilidad')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Данные, аналитика, ML, AI
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Данные, аналитика, ML, AI'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Данные, аналитика, ML, AI'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'conjunto'),
    (1, 'modelo'),
    (2, 'algoritmo'),
    (3, 'predicción'),
    (4, 'clasificación'),
    (5, 'procesar'),
    (6, 'varianza'),
    (7, 'covarianza'),
    (8, 'matriz'),
    (9, 'vector'),
    (10, 'distribución'),
    (11, 'normalización'),
    (12, 'agrupamiento'),
    (13, 'parámetro'),
    (14, 'estimador'),
    (15, 'muestreo'),
    (16, 'etiquetado'),
    (17, 'sobreajuste'),
    (18, 'predictor'),
    (19, 'predictivo'),
    (20, 'vectorización'),
    (21, 'incrustación'),
    (22, 'clasificador'),
    (23, 'clusterización'),
    (24, 'dimensionalidad'),
    (25, 'hiperparámetro'),
    (26, 'optimizador'),
    (27, 'convergencia'),
    (28, 'validable'),
    (29, 'entrenable'),
    (30, 'inferencial'),
    (31, 'probabilístico'),
    (32, 'neuronal'),
    (33, 'supervisado'),
    (34, 'tokenización'),
    (35, 'estadística'),
    (36, 'tokenizado'),
    (37, 'vectorial'),
    (38, 'probabilística'),
    (39, 'bayesiano'),
    (40, 'inferidor'),
    (41, 'predictibilidad'),
    (42, 'generativo'),
    (43, 'vectorizador'),
    (44, 'lingüístico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Безопасность, приватность, риски
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Безопасность, приватность, риски'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Безопасность, приватность, риски'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ataque'),
    (1, 'filtración'),
    (2, 'auditoría'),
    (3, 'intrusión'),
    (4, 'anonimización'),
    (5, 'brecha'),
    (6, 'cortafuegos'),
    (7, 'ciberseguridad'),
    (8, 'encriptación'),
    (9, 'encriptar'),
    (10, 'descifrar'),
    (11, 'atacante'),
    (12, 'explotable'),
    (13, 'explotación'),
    (14, 'parcheado'),
    (15, 'parchear'),
    (16, 'securizar'),
    (17, 'securización'),
    (18, 'trazable'),
    (19, 'auditable'),
    (20, 'auditabilidad'),
    (21, 'normativo'),
    (22, 'seudonimización'),
    (23, 'anonimizar'),
    (24, 'credencialización'),
    (25, 'cortafuego'),
    (26, 'antivirus'),
    (27, 'antimalware'),
    (28, 'intrusivo'),
    (29, 'malicioso'),
    (30, 'securizado'),
    (31, 'verificable'),
    (32, 'verificabilidad'),
    (33, 'auditado'),
    (34, 'trazado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Инфраструктура, cloud, DevOps
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Инфраструктура, cloud, DevOps'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Инфраструктура, cloud, DevOps'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'clúster'),
    (1, 'monitorización'),
    (2, 'escalado'),
    (3, 'automatizar'),
    (4, 'balanceador'),
    (5, 'orquestación'),
    (6, 'contenedorización'),
    (7, 'telemetría'),
    (8, 'observabilidad'),
    (9, 'latencia'),
    (10, 'nodo'),
    (11, 'aprovisionar'),
    (12, 'respaldar'),
    (13, 'virtualización'),
    (14, 'orquestador'),
    (15, 'aprovisionamiento'),
    (16, 'contenedorizado'),
    (17, 'continuo'),
    (18, 'balanceo'),
    (19, 'redundancia'),
    (20, 'tolerancia'),
    (21, 'monitorizar'),
    (22, 'provisionar'),
    (23, 'contenerizar'),
    (24, 'virtualizar'),
    (25, 'escalador'),
    (26, 'replicación'),
    (27, 'orquestable'),
    (28, 'aprovisionable'),
    (29, 'monitorizable'),
    (30, 'redundante'),
    (31, 'balanceado'),
    (32, 'virtualizado'),
    (33, 'contenerizado'),
    (34, 'automatizable'),
    (35, 'desplegador'),
    (36, 'despliegable'),
    (37, 'elasticidad'),
    (38, 'concurrencia'),
    (39, 'persistencia')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Наука: методы, исследования, гипотезы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Наука: методы, исследования, гипотезы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Наука: методы, исследования, гипотезы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'experimento'),
    (1, 'teoría'),
    (2, 'medición'),
    (3, 'laboratorio'),
    (4, 'replicabilidad'),
    (5, 'reproducibilidad'),
    (6, 'validez'),
    (7, 'protocolo'),
    (8, 'metodología'),
    (9, 'ensayo'),
    (10, 'controlado'),
    (11, 'aleatorio'),
    (12, 'falsabilidad'),
    (13, 'experimental'),
    (14, 'longitudinal'),
    (15, 'cuantitativo'),
    (16, 'cualitativo'),
    (17, 'empírico'),
    (18, 'teórico'),
    (19, 'metodológico'),
    (20, 'reproducible'),
    (21, 'replicable'),
    (22, 'falsable'),
    (23, 'aleatorización'),
    (24, 'placebo'),
    (25, 'ciego'),
    (26, 'metaanálisis'),
    (27, 'independiente'),
    (28, 'experimentalidad'),
    (29, 'operacionalización'),
    (30, 'diseño'),
    (31, 'contrastación'),
    (32, 'muestreal'),
    (33, 'bibliográfica'),
    (34, 'científica'),
    (35, 'triangulación'),
    (36, 'cohortal'),
    (37, 'exploratorio'),
    (38, 'confirmatorio'),
    (39, 'explicativo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Продукты, UX, метрики
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Продукты, UX, метрики'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Продукты, UX, метрики'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'iterar'),
    (1, 'accesibilidad'),
    (2, 'navegación'),
    (3, 'flujo'),
    (4, 'prototipado'),
    (5, 'navegabilidad'),
    (6, 'usable'),
    (7, 'intuitivo'),
    (8, 'conversacional'),
    (9, 'fluidez'),
    (10, 'activación'),
    (11, 'heurística'),
    (12, 'testeo'),
    (13, 'microcopia'),
    (14, 'prototipable'),
    (15, 'navegable'),
    (16, 'intuitividad'),
    (17, 'activable'),
    (18, 'retenible'),
    (19, 'convertible'),
    (20, 'descubribilidad'),
    (21, 'adoptabilidad'),
    (22, 'prototipar'),
    (23, 'rediseñar'),
    (24, 'simplificar'),
    (25, 'experimentar'),
    (26, 'personalizar'),
    (27, 'internacionalizar'),
    (28, 'segmentar'),
    (29, 'activar'),
    (30, 'convertir'),
    (31, 'instrumentar'),
    (32, 'instrumentación'),
    (33, 'navegacional'),
    (34, 'persuasivo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Ошибки, инциденты, качество
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Ошибки, инциденты, качество'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Ошибки, инциденты, качество'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'severidad'),
    (1, 'estabilidad'),
    (2, 'degradación'),
    (3, 'reproducir'),
    (4, 'investigar'),
    (5, 'degradado'),
    (6, 'degradar'),
    (7, 'anomalía'),
    (8, 'observable'),
    (9, 'remediación'),
    (10, 'contención'),
    (11, 'parcheo'),
    (12, 'regresivo'),
    (13, 'criticidad'),
    (14, 'severo'),
    (15, 'intermitente'),
    (16, 'estable'),
    (17, 'monitoreo'),
    (18, 'incidentología'),
    (19, 'recuperable'),
    (20, 'fallar'),
    (21, 'degradable'),
    (22, 'diagnosticable'),
    (23, 'resolutivo'),
    (24, 'mitigador'),
    (25, 'correctivo'),
    (26, 'preventivo'),
    (27, 'estabilizador'),
    (28, 'estabilización'),
    (29, 'reproceso'),
    (30, 'reprocesar'),
    (31, 'reintento'),
    (32, 'reabrir'),
    (33, 'reapertura'),
    (34, 'reincidencia')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Технологии, наука, ИИ / Цифровое общество
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Цифровое общество'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Технологии, наука, ИИ'
    AND ws.title = 'Цифровое общество'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'información'),
    (1, 'inclusión'),
    (2, 'exclusión'),
    (3, 'ciudadanía'),
    (4, 'alfabetización'),
    (5, 'electrónico'),
    (6, 'cibersociedad'),
    (7, 'conectividad'),
    (8, 'algorítmico'),
    (9, 'tecnosociedad'),
    (10, 'cibercultura'),
    (11, 'desigualdad'),
    (12, 'informacional'),
    (13, 'ciberespacio'),
    (14, 'ciberciudadanía'),
    (15, 'digitalizar'),
    (16, 'informatizar'),
    (17, 'informatizado'),
    (18, 'teleasistencia'),
    (19, 'telemedicina'),
    (20, 'teledemocracia'),
    (21, 'tecnopolítica'),
    (22, 'tecnopolítico'),
    (23, 'hiperconectado'),
    (24, 'hiperconectividad')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Государство и институты
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Государство и институты'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Государство и институты'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'estado'),
    (1, 'institución'),
    (2, 'ministerio'),
    (3, 'parlamento'),
    (4, 'gabinete'),
    (5, 'presidencia'),
    (6, 'legislatura'),
    (7, 'ejecutivo'),
    (8, 'legislativo'),
    (9, 'judicial'),
    (10, 'constitucional'),
    (11, 'constitución'),
    (12, 'soberanía'),
    (13, 'descentralización'),
    (14, 'centralización'),
    (15, 'atribución'),
    (16, 'mandato'),
    (17, 'regulación'),
    (18, 'organismo'),
    (19, 'autoridad'),
    (20, 'funcionario'),
    (21, 'institucionalidad'),
    (22, 'gobernabilidad'),
    (23, 'estatal'),
    (24, 'republicano'),
    (25, 'federal'),
    (26, 'unitario'),
    (27, 'municipal'),
    (28, 'regional'),
    (29, 'provincial'),
    (30, 'autonómico'),
    (31, 'ministerial'),
    (32, 'parlamentario'),
    (33, 'gubernamental'),
    (34, 'jurisdiccional'),
    (35, 'regulador'),
    (36, 'regulatorio'),
    (37, 'centralizado'),
    (38, 'interinstitucional'),
    (39, 'potestad')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Выборы, партии, общественное мнение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Выборы, партии, общественное мнение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Выборы, партии, общественное мнение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'voto'),
    (1, 'votante'),
    (2, 'urna'),
    (3, 'sondeo'),
    (4, 'electorado'),
    (5, 'escrutinio'),
    (6, 'papeleta'),
    (7, 'circunscripción'),
    (8, 'abstención'),
    (9, 'abstencionismo'),
    (10, 'propaganda'),
    (11, 'mitin'),
    (12, 'coalición'),
    (13, 'oficialismo'),
    (14, 'pluralismo'),
    (15, 'ideología'),
    (16, 'conservador'),
    (17, 'progresista'),
    (18, 'liberal'),
    (19, 'socialista'),
    (20, 'centrista'),
    (21, 'populista'),
    (22, 'electoral'),
    (23, 'elector'),
    (24, 'partidario'),
    (25, 'bipartidismo'),
    (26, 'multipartidismo'),
    (27, 'primarias'),
    (28, 'balotaje'),
    (29, 'sufragio'),
    (30, 'referendo'),
    (31, 'plebiscito'),
    (32, 'presidenciable'),
    (33, 'disenso'),
    (34, 'mayoritario')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Право, суд, договоры
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Право, суд, договоры'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Право, суд, договоры'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'tribunal'),
    (1, 'juez'),
    (2, 'fiscal'),
    (3, 'sentencia'),
    (4, 'apelación'),
    (5, 'pena'),
    (6, 'acusación'),
    (7, 'defensa'),
    (8, 'dictamen'),
    (9, 'perito'),
    (10, 'escritura'),
    (11, 'procesal'),
    (12, 'mercantil'),
    (13, 'nulo'),
    (14, 'anulable'),
    (15, 'recurrible'),
    (16, 'apelable'),
    (17, 'condenatorio'),
    (18, 'absolutorio'),
    (19, 'cautelar'),
    (20, 'probatorio'),
    (21, 'probanza'),
    (22, 'prescripción'),
    (23, 'embargo'),
    (24, 'decomiso'),
    (25, 'allanamiento'),
    (26, 'homologación'),
    (27, 'ratificación'),
    (28, 'jurisprudencia'),
    (29, 'casación'),
    (30, 'suplicación'),
    (31, 'auto'),
    (32, 'sobreseimiento'),
    (33, 'imputación'),
    (34, 'imputar'),
    (35, 'absolver'),
    (36, 'condenar'),
    (37, 'demandado'),
    (38, 'querellante'),
    (39, 'querellado')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Миграция, гражданство, статус
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Миграция, гражданство, статус'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Миграция, гражданство, статус'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'migración'),
    (1, 'emigración'),
    (2, 'inmigración'),
    (3, 'inmigrante'),
    (4, 'emigrante'),
    (5, 'residencia'),
    (6, 'asilo'),
    (7, 'refugio'),
    (8, 'refugiado'),
    (9, 'naturalización'),
    (10, 'expulsión'),
    (11, 'diáspora'),
    (12, 'tránsito'),
    (13, 'embajada'),
    (14, 'residente'),
    (15, 'migratorio'),
    (16, 'asilado'),
    (17, 'documentado'),
    (18, 'indocumentado'),
    (19, 'nacionalizado'),
    (20, 'retornado'),
    (21, 'desplazado'),
    (22, 'reasentamiento'),
    (23, 'acogida'),
    (24, 'estatuto'),
    (25, 'admisión'),
    (26, 'permanencia'),
    (27, 'empadronamiento'),
    (28, 'nacionalizar'),
    (29, 'repatriar'),
    (30, 'deportar'),
    (31, 'expulsar'),
    (32, 'regularizar'),
    (33, 'fronterizo'),
    (34, 'transitar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Социальные группы и неравенство
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Социальные группы и неравенство'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Социальные группы и неравенство'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'estrato'),
    (1, 'minoría'),
    (2, 'mayoría'),
    (3, 'élite'),
    (4, 'discriminación'),
    (5, 'segregación'),
    (6, 'marginalidad'),
    (7, 'privilegio'),
    (8, 'precariedad'),
    (9, 'población'),
    (10, 'demografía'),
    (11, 'generación'),
    (12, 'etnia'),
    (13, 'raza'),
    (14, 'discapacidad'),
    (15, 'diversidad'),
    (16, 'interseccionalidad'),
    (17, 'estigma'),
    (18, 'segregado'),
    (19, 'marginal'),
    (20, 'vulnerable'),
    (21, 'privilegiado'),
    (22, 'precario'),
    (23, 'excluido'),
    (24, 'minoritario'),
    (25, 'identitario'),
    (26, 'demográfico'),
    (27, 'generacional'),
    (28, 'étnico'),
    (29, 'racial'),
    (30, 'redistribución'),
    (31, 'redistributivo'),
    (32, 'inequidad'),
    (33, 'vulnerabilizar'),
    (34, 'marginar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Преступность и безопасность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Преступность и безопасность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Преступность и безопасность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'crimen'),
    (1, 'delincuencia'),
    (2, 'fraude'),
    (3, 'corrupción'),
    (4, 'homicidio'),
    (5, 'violencia'),
    (6, 'agresor'),
    (7, 'detenido'),
    (8, 'patrulla'),
    (9, 'condena'),
    (10, 'prisión'),
    (11, 'cárcel'),
    (12, 'impunidad'),
    (13, 'persecución'),
    (14, 'incautación'),
    (15, 'pandilla'),
    (16, 'extorsión'),
    (17, 'secuestro'),
    (18, 'contrabando'),
    (19, 'blanqueo'),
    (20, 'soborno'),
    (21, 'cohecho'),
    (22, 'intimidación'),
    (23, 'delictivo'),
    (24, 'criminal'),
    (25, 'punitivo'),
    (26, 'carcelario'),
    (27, 'policial'),
    (28, 'agravante'),
    (29, 'flagrante'),
    (30, 'procesado'),
    (31, 'recluso'),
    (32, 'prófugo'),
    (33, 'victimario'),
    (34, 'victimización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Международные отношения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Международные отношения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Международные отношения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'diplomacia'),
    (1, 'tratado'),
    (2, 'cumbre'),
    (3, 'cancillería'),
    (4, 'embajador'),
    (5, 'bilateral'),
    (6, 'multilateral'),
    (7, 'unilateral'),
    (8, 'geopolítica'),
    (9, 'geoestrategia'),
    (10, 'bloque'),
    (11, 'adhesión'),
    (12, 'membresía'),
    (13, 'tregua'),
    (14, 'armisticio'),
    (15, 'diplomático'),
    (16, 'interestatal'),
    (17, 'intergubernamental'),
    (18, 'supranacional'),
    (19, 'humanitario'),
    (20, 'pacificador'),
    (21, 'pacificación'),
    (22, 'enviado'),
    (23, 'delegatorio'),
    (24, 'canciller'),
    (25, 'acreditado'),
    (26, 'plenipotenciario'),
    (27, 'extradición'),
    (28, 'extraditar'),
    (29, 'cooperante'),
    (30, 'multilateralismo'),
    (31, 'unilateralismo'),
    (32, 'bilateralismo'),
    (33, 'regionalismo'),
    (34, 'supranacionalidad'),
    (35, 'interdependiente'),
    (36, 'geopolítico'),
    (37, 'geoestratégico'),
    (38, 'limítrofe'),
    (39, 'marítimo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Политика, общество, право / Общественные услуги и бюрократия
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Общественные услуги и бюрократия'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Политика, общество, право'
    AND ws.title = 'Общественные услуги и бюрократия'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'cola'),
    (1, 'burocrático'),
    (2, 'subsidio'),
    (3, 'prestación'),
    (4, 'impuesto'),
    (5, 'fiscalización'),
    (6, 'presencialidad'),
    (7, 'teletrámite'),
    (8, 'compulsado'),
    (9, 'certificador'),
    (10, 'licitación'),
    (11, 'pliego'),
    (12, 'concesionario'),
    (13, 'instanciar'),
    (14, 'tramitador'),
    (15, 'tramitado'),
    (16, 'archivístico'),
    (17, 'registral'),
    (18, 'catastro'),
    (19, 'catastral'),
    (20, 'empadronar'),
    (21, 'empadronado'),
    (22, 'censal'),
    (23, 'censo'),
    (24, 'certificante'),
    (25, 'certificable'),
    (26, 'compulsa'),
    (27, 'requerir'),
    (28, 'notificante'),
    (29, 'resolutorio'),
    (30, 'sancionador'),
    (31, 'concesional'),
    (32, 'concesionar'),
    (33, 'licitador'),
    (34, 'licitante'),
    (35, 'adjudicador'),
    (36, 'adjudicar'),
    (37, 'adjudicable'),
    (38, 'licitatorio'),
    (39, 'contratista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Макроэкономика: минимум
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Макроэкономика: минимум'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Макроэкономика: минимум'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'macroeconomía'),
    (1, 'recesión'),
    (2, 'inflación'),
    (3, 'deflación'),
    (4, 'desempleo'),
    (5, 'empleo'),
    (6, 'consumo'),
    (7, 'importación'),
    (8, 'balanza'),
    (9, 'superávit'),
    (10, 'ciclo'),
    (11, 'coyuntura'),
    (12, 'monetario'),
    (13, 'cambiario'),
    (14, 'arancelario'),
    (15, 'industrial'),
    (16, 'agregado'),
    (17, 'recesivo'),
    (18, 'expansivo'),
    (19, 'inflacionario'),
    (20, 'deflacionario'),
    (21, 'improductivo'),
    (22, 'estanflación'),
    (23, 'endeudamiento'),
    (24, 'insolvencia'),
    (25, 'iliquidez'),
    (26, 'competitividad'),
    (27, 'macroeconómico'),
    (28, 'microeconómico'),
    (29, 'superavitario'),
    (30, 'deficitario'),
    (31, 'presupuestario'),
    (32, 'contracíclico'),
    (33, 'procíclico'),
    (34, 'anticíclico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Личные финансы и инвестиции
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Личные финансы и инвестиции'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Личные финансы и инвестиции'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'rentabilidad'),
    (1, 'dividendo'),
    (2, 'acción'),
    (3, 'bono'),
    (4, 'fondo'),
    (5, 'capital'),
    (6, 'pasivo'),
    (7, 'hipoteca'),
    (8, 'pensión'),
    (9, 'jubilación'),
    (10, 'aportación'),
    (11, 'diversificación'),
    (12, 'diversificar'),
    (13, 'refinanciar'),
    (14, 'amortizar'),
    (15, 'amortización'),
    (16, 'compuesto'),
    (17, 'volatilidad'),
    (18, 'agresivo'),
    (19, 'moderado'),
    (20, 'especulativo'),
    (21, 'ilíquido'),
    (22, 'solvente'),
    (23, 'insolvente'),
    (24, 'bursátil'),
    (25, 'accionista'),
    (26, 'bonista'),
    (27, 'inversor'),
    (28, 'ahorrador'),
    (29, 'apalancamiento'),
    (30, 'apalancar'),
    (31, 'desinvertir'),
    (32, 'desinversión'),
    (33, 'revalorizar'),
    (34, 'revalorización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Недвижимость, ипотека, аренда
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Недвижимость, ипотека, аренда'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Недвижимость, ипотека, аренда'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'propiedad'),
    (1, 'tasación'),
    (2, 'parcela'),
    (3, 'notaría'),
    (4, 'condominio'),
    (5, 'inmobiliario'),
    (6, 'hipotecario'),
    (7, 'arrendaticio'),
    (8, 'habitable'),
    (9, 'inhabitable'),
    (10, 'amueblado'),
    (11, 'reformado'),
    (12, 'tasado'),
    (13, 'escriturado'),
    (14, 'hipotecar'),
    (15, 'subarrendar'),
    (16, 'desalojar'),
    (17, 'desalojo'),
    (18, 'okupación'),
    (19, 'ocupación'),
    (20, 'avalista'),
    (21, 'plusvalía'),
    (22, 'derribar'),
    (23, 'edificar'),
    (24, 'edificable'),
    (25, 'urbanizable'),
    (26, 'proindiviso'),
    (27, 'hipotecable'),
    (28, 'hipotecado'),
    (29, 'subarrendador'),
    (30, 'subarrendatario'),
    (31, 'desalojable'),
    (32, 'ocupacional'),
    (33, 'okupa'),
    (34, 'desahuciar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Налоги, зарплата, соцвзносы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Налоги, зарплата, соцвзносы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Налоги, зарплата, соцвзносы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'tributo'),
    (1, 'contribución'),
    (2, 'tramo'),
    (3, 'mínimo'),
    (4, 'pagador'),
    (5, 'empleador'),
    (6, 'asalariado'),
    (7, 'exento'),
    (8, 'declarable'),
    (9, 'tributable'),
    (10, 'recaudatorio'),
    (11, 'tributario'),
    (12, 'previsional'),
    (13, 'devengo'),
    (14, 'devengar'),
    (15, 'liquidar'),
    (16, 'cotizar'),
    (17, 'aportar'),
    (18, 'desgravar'),
    (19, 'desgravación'),
    (20, 'impositivo'),
    (21, 'imponibilidad'),
    (22, 'retenedor'),
    (23, 'retenido'),
    (24, 'deducibilidad'),
    (25, 'cotizante'),
    (26, 'contributivo'),
    (27, 'autoliquidación'),
    (28, 'autoliquidar'),
    (29, 'declarativo'),
    (30, 'declarante'),
    (31, 'fiscalizador'),
    (32, 'fiscalizado'),
    (33, 'inspeccionable'),
    (34, 'inspeccionar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Банки, кредиты, страхование
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Банки, кредиты, страхование'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Банки, кредиты, страхование'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'impago'),
    (1, 'bancario'),
    (2, 'crediticio'),
    (3, 'asegurador'),
    (4, 'asegurable'),
    (5, 'impagado'),
    (6, 'refinanciación'),
    (7, 'garante'),
    (8, 'garantizado'),
    (9, 'descubierto'),
    (10, 'coaseguro'),
    (11, 'reaseguro'),
    (12, 'financiero'),
    (13, 'bancarizar'),
    (14, 'bancarización'),
    (15, 'bancarizado'),
    (16, 'bancarrota'),
    (17, 'prestamista'),
    (18, 'prestatario'),
    (19, 'acreedor'),
    (20, 'deudor'),
    (21, 'hipotecante'),
    (22, 'prendario'),
    (23, 'prenda'),
    (24, 'avalado'),
    (25, 'aseguramiento'),
    (26, 'asegurabilidad'),
    (27, 'siniestralidad'),
    (28, 'siniestral'),
    (29, 'cubierto'),
    (30, 'sobregiro'),
    (31, 'moratorio'),
    (32, 'amortizable'),
    (33, 'amortizado'),
    (34, 'reasegurador')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Бизнес-модели и ценообразование
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Бизнес-модели и ценообразование'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Бизнес-модели и ценообразование'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'negocio'),
    (1, 'costo'),
    (2, 'fijo'),
    (3, 'recurrente'),
    (4, 'básico'),
    (5, 'monetizable'),
    (6, 'paquetizar'),
    (7, 'subvencionar'),
    (8, 'subsidiar'),
    (9, 'financiarse'),
    (10, 'costear'),
    (11, 'preciar'),
    (12, 'tarificación'),
    (13, 'tarificar'),
    (14, 'tarifario'),
    (15, 'tarifado'),
    (16, 'costeable'),
    (17, 'costeado'),
    (18, 'costeo'),
    (19, 'paquetización'),
    (20, 'empaquetamiento'),
    (21, 'paquetizado'),
    (22, 'comisionable'),
    (23, 'recurrencia'),
    (24, 'cruzado'),
    (25, 'subvención'),
    (26, 'subsidización'),
    (27, 'financiador'),
    (28, 'financiado'),
    (29, 'facturador'),
    (30, 'cobrable'),
    (31, 'diferido'),
    (32, 'prorrateo'),
    (33, 'prorratear'),
    (34, 'descontar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экономика, финансы, недвижимость / Кризисы, инфляция, рынки
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Кризисы, инфляция, рынки'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экономика, финансы, недвижимость'
    AND ws.title = 'Кризисы, инфляция, рынки'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'depresión'),
    (1, 'burbuja'),
    (2, 'desplome'),
    (3, 'pánico'),
    (4, 'contagio'),
    (5, 'devaluación'),
    (6, 'depreciación'),
    (7, 'escasez'),
    (8, 'quiebra'),
    (9, 'moratoria'),
    (10, 'rescate'),
    (11, 'austeridad'),
    (12, 'especulación'),
    (13, 'especulador'),
    (14, 'repunte'),
    (15, 'sobrecalentamiento'),
    (16, 'enfriamiento'),
    (17, 'depresivo'),
    (18, 'volátil'),
    (19, 'bajista'),
    (20, 'alcista'),
    (21, 'sistémico'),
    (22, 'energético'),
    (23, 'alimentario'),
    (24, 'choque'),
    (25, 'perturbación'),
    (26, 'turbulencia'),
    (27, 'diferencial'),
    (28, 'inflacionista'),
    (29, 'desinflación'),
    (30, 'desinflacionario'),
    (31, 'hiperinflación'),
    (32, 'hiperinflacionario'),
    (33, 'devaluatorio'),
    (34, 'depreciatorio'),
    (35, 'apreciatorio'),
    (36, 'estanflacionario'),
    (37, 'recesionista'),
    (38, 'contraccionista'),
    (39, 'reestructuración')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Жанры и формы искусства
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Жанры и формы искусства'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Жанры и формы искусства'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'arte'),
    (1, 'forma'),
    (2, 'danza'),
    (3, 'poesía'),
    (4, 'dramaturgia'),
    (5, 'tragedia'),
    (6, 'animación'),
    (7, 'performance'),
    (8, 'grabado'),
    (9, 'collage'),
    (10, 'retrato'),
    (11, 'bodegón'),
    (12, 'ópera'),
    (13, 'ballet'),
    (14, 'sinfonía'),
    (15, 'sonata'),
    (16, 'coral'),
    (17, 'narrativo'),
    (18, 'lírico'),
    (19, 'dramático'),
    (20, 'pictórico'),
    (21, 'escultórico'),
    (22, 'cinematográfico'),
    (23, 'teatral'),
    (24, 'coreográfico'),
    (25, 'literario'),
    (26, 'poético'),
    (27, 'novelístico'),
    (28, 'ensayístico'),
    (29, 'documentalista'),
    (30, 'vanguardista'),
    (31, 'realista'),
    (32, 'simbolista'),
    (33, 'impresionista'),
    (34, 'expresionista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Кино, литература, критика
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Кино, литература, критика'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Кино, литература, критика'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'guionista'),
    (1, 'plano'),
    (2, 'encuadre'),
    (3, 'antagonista'),
    (4, 'diálogo'),
    (5, 'monólogo'),
    (6, 'subtexto'),
    (7, 'simbolismo'),
    (8, 'metáfora'),
    (9, 'alegoría'),
    (10, 'ironía'),
    (11, 'verosimilitud'),
    (12, 'intertextualidad'),
    (13, 'autoría'),
    (14, 'autoral'),
    (15, 'fílmico'),
    (16, 'estilístico'),
    (17, 'mediometraje'),
    (18, 'filmografía'),
    (19, 'filmar'),
    (20, 'rodaje'),
    (21, 'doblar'),
    (22, 'subtitulado'),
    (23, 'subtitular'),
    (24, 'montajista'),
    (25, 'escenografía'),
    (26, 'escenógrafo'),
    (27, 'secuencialidad'),
    (28, 'contrapicado'),
    (29, 'encuadrar'),
    (30, 'omnisciente'),
    (31, 'narratología'),
    (32, 'narratológico'),
    (33, 'diegético'),
    (34, 'extradiegético')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Журналистика и медиаформаты
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Журналистика и медиаформаты'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Журналистика и медиаформаты'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'periodismo'),
    (1, 'entradilla'),
    (2, 'edición'),
    (3, 'boletín'),
    (4, 'exclusiva'),
    (5, 'primicia'),
    (6, 'desmentido'),
    (7, 'rumor'),
    (8, 'viralidad'),
    (9, 'fotoperiodismo'),
    (10, 'infografía'),
    (11, 'gráfico'),
    (12, 'corresponsalía'),
    (13, 'mediático'),
    (14, 'periodístico'),
    (15, 'editorialista'),
    (16, 'opinativo'),
    (17, 'informativo'),
    (18, 'investigativo'),
    (19, 'sensacionalista'),
    (20, 'contrastable'),
    (21, 'redactor'),
    (22, 'reportero'),
    (23, 'cronista'),
    (24, 'entrevistador'),
    (25, 'entrevistado'),
    (26, 'columnista'),
    (27, 'articulista'),
    (28, 'editorializar'),
    (29, 'editorialización'),
    (30, 'titularizar'),
    (31, 'titularización'),
    (32, 'maquetación'),
    (33, 'maquetar'),
    (34, 'desinformación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / История и культурная память
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'История и культурная память'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'История и культурная память'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'conmemoración'),
    (1, 'testimonio'),
    (2, 'legado'),
    (3, 'tradición'),
    (4, 'genealogía'),
    (5, 'linaje'),
    (6, 'civilización'),
    (7, 'época'),
    (8, 'período'),
    (9, 'siglo'),
    (10, 'antigüedad'),
    (11, 'modernidad'),
    (12, 'contemporaneidad'),
    (13, 'revolución'),
    (14, 'guerra'),
    (15, 'posguerra'),
    (16, 'exilio'),
    (17, 'dictadura'),
    (18, 'democracia'),
    (19, 'colonización'),
    (20, 'descolonización'),
    (21, 'historiografía'),
    (22, 'historiador'),
    (23, 'arqueología'),
    (24, 'arqueológico'),
    (25, 'patrimonial'),
    (26, 'conmemorativo'),
    (27, 'memorial'),
    (28, 'testimonial'),
    (29, 'ancestral'),
    (30, 'fundacional'),
    (31, 'memorialismo'),
    (32, 'memorialista'),
    (33, 'historicidad'),
    (34, 'historicista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Традиции, праздники, идентичность
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Традиции, праздники, идентичность'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Традиции, праздники, идентичность'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ritual'),
    (1, 'ceremonia'),
    (2, 'símbolo'),
    (3, 'emblema'),
    (4, 'folclore'),
    (5, 'procesión'),
    (6, 'carnaval'),
    (7, 'festividad'),
    (8, 'herencia'),
    (9, 'autóctono'),
    (10, 'ceremonial'),
    (11, 'festivo'),
    (12, 'tradicional'),
    (13, 'ritualizado'),
    (14, 'mestizaje'),
    (15, 'arraigado'),
    (16, 'pertenecer'),
    (17, 'conmemorar'),
    (18, 'transmitir'),
    (19, 'preservar'),
    (20, 'ritualidad'),
    (21, 'ritualismo'),
    (22, 'celebrante'),
    (23, 'festejo'),
    (24, 'festejar'),
    (25, 'conmemorador'),
    (26, 'procesional'),
    (27, 'carnavalesco'),
    (28, 'folclórico'),
    (29, 'folclorista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Рецензии и интерпретация
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Рецензии и интерпретация'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Рецензии и интерпретация'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'composición'),
    (1, 'estructura'),
    (2, 'originalidad'),
    (3, 'influencia'),
    (4, 'contextualización'),
    (5, 'criticar'),
    (6, 'elogiar'),
    (7, 'insinuar'),
    (8, 'contrastar'),
    (9, 'comparativo'),
    (10, 'valorativo'),
    (11, 'irónico'),
    (12, 'ambiguo'),
    (13, 'sugerente'),
    (14, 'exégesis'),
    (15, 'relectura'),
    (16, 'comentador'),
    (17, 'reinterpretación'),
    (18, 'reinterpretar'),
    (19, 'polisemia'),
    (20, 'polisémico'),
    (21, 'ambivalente'),
    (22, 'connotativo'),
    (23, 'denotación'),
    (24, 'denotativo'),
    (25, 'matización'),
    (26, 'problematización'),
    (27, 'problematizador'),
    (28, 'contraargumentar'),
    (29, 'recepcional'),
    (30, 'canonización'),
    (31, 'canonizar'),
    (32, 'contextualizador'),
    (33, 'contextualizado'),
    (34, 'intertextual'),
    (35, 'simbolización'),
    (36, 'simbolizar'),
    (37, 'metafórico'),
    (38, 'alegórico'),
    (39, 'subtextual')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Культура, медиа, литература / Культурные индустрии
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Культурные индустрии'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Культура, медиа, литература'
    AND ws.title = 'Культурные индустрии'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'industria'),
    (1, 'productora'),
    (2, 'distribuidora'),
    (3, 'exhibición'),
    (4, 'mecenazgo'),
    (5, 'curaduría'),
    (6, 'programación'),
    (7, 'repertorio'),
    (8, 'taquilla'),
    (9, 'producción'),
    (10, 'curador'),
    (11, 'gestor'),
    (12, 'productor'),
    (13, 'exhibidor'),
    (14, 'promotor'),
    (15, 'patrocinable'),
    (16, 'audiovisual'),
    (17, 'escénico'),
    (18, 'musical'),
    (19, 'museístico'),
    (20, 'curatorial'),
    (21, 'industrialización'),
    (22, 'industrializado'),
    (23, 'culturalización'),
    (24, 'creatividad'),
    (25, 'editorialidad'),
    (26, 'editabilidad'),
    (27, 'publicable'),
    (28, 'exhibible'),
    (29, 'programable'),
    (30, 'festivalero'),
    (31, 'festivalesco'),
    (32, 'curable'),
    (33, 'museal'),
    (34, 'musealización'),
    (35, 'musealizar'),
    (36, 'escenificación'),
    (37, 'escenificar'),
    (38, 'taquillaje'),
    (39, 'taquillero')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Эмоции и внутренние состояния B2
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Эмоции и внутренние состояния B2'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Эмоции и внутренние состояния B2'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'melancolía'),
    (1, 'apatía'),
    (2, 'autoestima'),
    (3, 'inseguridad'),
    (4, 'desconfianza'),
    (5, 'saturación'),
    (6, 'deseo'),
    (7, 'rechazo'),
    (8, 'desapego'),
    (9, 'duelo'),
    (10, 'trauma'),
    (11, 'resiliencia'),
    (12, 'autorregulación'),
    (13, 'introspección'),
    (14, 'introspectivo'),
    (15, 'ansioso'),
    (16, 'angustiado'),
    (17, 'eufórico'),
    (18, 'frustrado'),
    (19, 'resentido'),
    (20, 'culpable'),
    (21, 'avergonzado'),
    (22, 'aliviado'),
    (23, 'nostálgico'),
    (24, 'melancólico'),
    (25, 'apático'),
    (26, 'irritable'),
    (27, 'sereno'),
    (28, 'agotado'),
    (29, 'insegurizar'),
    (30, 'desregular'),
    (31, 'desregulación'),
    (32, 'somatización'),
    (33, 'somatizar'),
    (34, 'disociación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Коммуникация и границы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Коммуникация и границы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Коммуникация и границы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'consentimiento'),
    (1, 'empatía'),
    (2, 'asertividad'),
    (3, 'invalidación'),
    (4, 'manipulación'),
    (5, 'reciprocidad'),
    (6, 'expectativa'),
    (7, 'negativa'),
    (8, 'indisponibilidad'),
    (9, 'explícito'),
    (10, 'implícito'),
    (11, 'asertivo'),
    (12, 'invasivo'),
    (13, 'manipulador'),
    (14, 'invalidar'),
    (15, 'presionar'),
    (16, 'acercarse'),
    (17, 'limitar'),
    (18, 'delimitar'),
    (19, 'delimitación'),
    (20, 'consensuado'),
    (21, 'comunicabilidad'),
    (22, 'comunicativo'),
    (23, 'escuchante'),
    (24, 'empatizar'),
    (25, 'disculpable'),
    (26, 'reparador'),
    (27, 'reprochable'),
    (28, 'invalidante'),
    (29, 'manipular')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Конфликты, манипуляции, доверие
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Конфликты, манипуляции, доверие'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Конфликты, манипуляции, доверие'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'chantaje'),
    (1, 'traición'),
    (2, 'abuso'),
    (3, 'dominio'),
    (4, 'sumisión'),
    (5, 'agresividad'),
    (6, 'evasión'),
    (7, 'rencor'),
    (8, 'intimidar'),
    (9, 'amenazar'),
    (10, 'desconfiar'),
    (11, 'sospechar'),
    (12, 'traicionar'),
    (13, 'culpabilizar'),
    (14, 'victimizar'),
    (15, 'coerción'),
    (16, 'coercitivo'),
    (17, 'abusivo'),
    (18, 'controlador'),
    (19, 'dominante'),
    (20, 'sumiso'),
    (21, 'hostil'),
    (22, 'defensivo'),
    (23, 'coaccionar'),
    (24, 'coactivo'),
    (25, 'intimidatorio'),
    (26, 'amenazante'),
    (27, 'manipulativo'),
    (28, 'deslegitimar'),
    (29, 'deslegitimación'),
    (30, 'desacreditar'),
    (31, 'desacreditación'),
    (32, 'victimismo'),
    (33, 'victimista'),
    (34, 'culpabilización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Семья, партнёрство, сообщество
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Семья, партнёрство, сообщество'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Семья, партнёрство, сообщество'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'crianza'),
    (1, 'maternidad'),
    (2, 'paternidad'),
    (3, 'hogar'),
    (4, 'divorcio'),
    (5, 'adopción'),
    (6, 'padrastro'),
    (7, 'madrastra'),
    (8, 'hijastro'),
    (9, 'suegro'),
    (10, 'yerno'),
    (11, 'nuera'),
    (12, 'cuñado'),
    (13, 'coparentalidad'),
    (14, 'coparental'),
    (15, 'cuidador'),
    (16, 'cuidadora'),
    (17, 'conyugal'),
    (18, 'parental'),
    (19, 'fraternal'),
    (20, 'solidario'),
    (21, 'nupcial'),
    (22, 'conyugalidad'),
    (23, 'parentela'),
    (24, 'filiación'),
    (25, 'filial'),
    (26, 'consanguinidad'),
    (27, 'consanguíneo'),
    (28, 'corresidencia'),
    (29, 'convivencial')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Личность, привычки, поведение
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Личность, привычки, поведение'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Личность, привычки, поведение'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'temperamento'),
    (1, 'conducta'),
    (2, 'comportamiento'),
    (3, 'procrastinación'),
    (4, 'perfeccionismo'),
    (5, 'impulsividad'),
    (6, 'introversión'),
    (7, 'extraversión'),
    (8, 'apertura'),
    (9, 'neuroticismo'),
    (10, 'rasgo'),
    (11, 'autoimagen'),
    (12, 'autoconcepto'),
    (13, 'autocrítica'),
    (14, 'autoexigencia'),
    (15, 'autoobservación'),
    (16, 'repetición'),
    (17, 'automatismo'),
    (18, 'compulsión'),
    (19, 'compulsivo'),
    (20, 'rígido'),
    (21, 'constante'),
    (22, 'perseverante'),
    (23, 'disciplinado'),
    (24, 'habituarse'),
    (25, 'deshabituarse'),
    (26, 'adaptarse'),
    (27, 'conductual'),
    (28, 'temperamental'),
    (29, 'caracterológico'),
    (30, 'extravertido'),
    (31, 'abierto'),
    (32, 'neurótico'),
    (33, 'autoeficacia'),
    (34, 'autoconocimiento')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Психология, отношения, социальное взаимодействие / Стресс, выгорание, баланс
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Стресс, выгорание, баланс'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Психология, отношения, социальное взаимодействие'
    AND ws.title = 'Стресс, выгорание, баланс'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'desgaste'),
    (1, 'insomnio'),
    (2, 'irritabilidad'),
    (3, 'meditación'),
    (4, 'bienestar'),
    (5, 'malestar'),
    (6, 'fatiga'),
    (7, 'hiperactividad'),
    (8, 'crónico'),
    (9, 'agudo'),
    (10, 'quemarse'),
    (11, 'regularse'),
    (12, 'meditar'),
    (13, 'sobrecargado'),
    (14, 'exhausto'),
    (15, 'fatigado'),
    (16, 'desbordado'),
    (17, 'estresor'),
    (18, 'estresante'),
    (19, 'estresarse'),
    (20, 'sobrecargar'),
    (21, 'extenuación'),
    (22, 'extenuado'),
    (23, 'extenuar'),
    (24, 'fatigar'),
    (25, 'hiperexigencia'),
    (26, 'relajante'),
    (27, 'dosificación'),
    (28, 'balancear'),
    (29, 'pausado'),
    (30, 'pausar'),
    (31, 'insomne'),
    (32, 'irritar'),
    (33, 'recuperativo'),
    (34, 'restaurativo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Система здравоохранения
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Система здравоохранения'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Система здравоохранения'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sanidad'),
    (1, 'urgencias'),
    (2, 'enfermera'),
    (3, 'vacunación'),
    (4, 'cribado'),
    (5, 'epidemiología'),
    (6, 'epidemiológico'),
    (7, 'asistencial'),
    (8, 'hospitalario'),
    (9, 'clínico'),
    (10, 'hospitalizar'),
    (11, 'sanitarista'),
    (12, 'hospitalización'),
    (13, 'mutualidad'),
    (14, 'ambulatorización'),
    (15, 'vacunatorio'),
    (16, 'vacunador'),
    (17, 'inmunización'),
    (18, 'cribador'),
    (19, 'cribaje'),
    (20, 'asistencialidad'),
    (21, 'derivable'),
    (22, 'hospitalizable'),
    (23, 'urgenciólogo'),
    (24, 'facultativo'),
    (25, 'sanitarización'),
    (26, 'sanitarizar'),
    (27, 'derivador'),
    (28, 'ambulatorial'),
    (29, 'ambulatorizar'),
    (30, 'hospitalista'),
    (31, 'hospitalizado'),
    (32, 'ingresable'),
    (33, 'ingresado'),
    (34, 'admisionista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Диагностика, лечение, обследования
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Диагностика, лечение, обследования'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Диагностика, лечение, обследования'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'exploración'),
    (1, 'radiografía'),
    (2, 'ecografía'),
    (3, 'resonancia'),
    (4, 'tomografía'),
    (5, 'biopsia'),
    (6, 'endoscopia'),
    (7, 'electrocardiograma'),
    (8, 'signo'),
    (9, 'terapia'),
    (10, 'medicación'),
    (11, 'cirugía'),
    (12, 'anestesia'),
    (13, 'rehabilitación'),
    (14, 'fisioterapia'),
    (15, 'tratar'),
    (16, 'medicar'),
    (17, 'operar'),
    (18, 'rehabilitar'),
    (19, 'analítico'),
    (20, 'radiológico'),
    (21, 'quirúrgico'),
    (22, 'terapéutico'),
    (23, 'anamnesis'),
    (24, 'palpación'),
    (25, 'sintomatología'),
    (26, 'sintomático'),
    (27, 'asintomático'),
    (28, 'radiografiar'),
    (29, 'ecografiar'),
    (30, 'biopsiar'),
    (31, 'endoscópico'),
    (32, 'tomográfico'),
    (33, 'resonador'),
    (34, 'electrocardiográfico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Хронические состояния и профилактика
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Хронические состояния и профилактика'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Хронические состояния и профилактика'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sedentarismo'),
    (1, 'obesidad'),
    (2, 'hipertensión'),
    (3, 'diabetes'),
    (4, 'colesterol'),
    (5, 'tabaquismo'),
    (6, 'alcoholismo'),
    (7, 'adherencia'),
    (8, 'recaída'),
    (9, 'complicación'),
    (10, 'brote'),
    (11, 'comorbilidad'),
    (12, 'abandonar'),
    (13, 'mantener'),
    (14, 'estabilizar'),
    (15, 'controlable'),
    (16, 'estabilizado'),
    (17, 'degenerativo'),
    (18, 'metabólico'),
    (19, 'cardiovascular'),
    (20, 'respiratorio'),
    (21, 'cronicidad'),
    (22, 'cronificación'),
    (23, 'cronificar'),
    (24, 'sedentario'),
    (25, 'hipertenso'),
    (26, 'diabético'),
    (27, 'hipercolesterolemia'),
    (28, 'dislipidemia'),
    (29, 'tabaquista')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Питание, сон, спорт
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Питание, сон, спорт'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Питание, сон, спорт'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'nutrición'),
    (1, 'alimentación'),
    (2, 'proteína'),
    (3, 'carbohidrato'),
    (4, 'grasa'),
    (5, 'fibra'),
    (6, 'vitamina'),
    (7, 'mineral'),
    (8, 'hidratación'),
    (9, 'fuerza'),
    (10, 'deportista'),
    (11, 'caloría'),
    (12, 'metabolismo'),
    (13, 'digestión'),
    (14, 'saciedad'),
    (15, 'ayuno'),
    (16, 'suplemento'),
    (17, 'suplementación'),
    (18, 'muscular'),
    (19, 'aeróbico'),
    (20, 'anaeróbico'),
    (21, 'nutritivo'),
    (22, 'calórico'),
    (23, 'digestivo'),
    (24, 'hidratado'),
    (25, 'deshidratado'),
    (26, 'estirar'),
    (27, 'fortalecer'),
    (28, 'hidratarse'),
    (29, 'alimentarse')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Медицина и технологии
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Медицина и технологии'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Медицина и технологии'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'teleconsulta'),
    (1, 'telesalud'),
    (2, 'biosensor'),
    (3, 'implante'),
    (4, 'prótesis'),
    (5, 'robótica'),
    (6, 'radiología'),
    (7, 'videoconsulta'),
    (8, 'genómica'),
    (9, 'genómico'),
    (10, 'bioinformática'),
    (11, 'bioinformático'),
    (12, 'nanotecnología'),
    (13, 'nanomedicina'),
    (14, 'conectado'),
    (15, 'implantable'),
    (16, 'protésico'),
    (17, 'robotizado'),
    (18, 'telemédico'),
    (19, 'teleasistencial'),
    (20, 'telediagnóstico'),
    (21, 'telediagnosticar'),
    (22, 'telemonitorización'),
    (23, 'telemonitorizar'),
    (24, 'telemática'),
    (25, 'robotización'),
    (26, 'robotizar'),
    (27, 'robótico'),
    (28, 'protetización'),
    (29, 'protetizar'),
    (30, 'implantología'),
    (31, 'implantólogo'),
    (32, 'biosensórica'),
    (33, 'biosensorización'),
    (34, 'sensorización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Здоровье, медицина, образ жизни / Риски, побочные эффекты, инструкции
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Риски, побочные эффекты, инструкции'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Здоровье, медицина, образ жизни'
    AND ws.title = 'Риски, побочные эффекты, инструкции'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'advertencia'),
    (1, 'sobredosis'),
    (2, 'intolerancia'),
    (3, 'toxicidad'),
    (4, 'toxicología'),
    (5, 'abstinencia'),
    (6, 'somnolencia'),
    (7, 'erupción'),
    (8, 'farmacovigilancia'),
    (9, 'indicación'),
    (10, 'suspensión'),
    (11, 'aumentar'),
    (12, 'adverso'),
    (13, 'contraindicado'),
    (14, 'alérgico'),
    (15, 'tóxico'),
    (16, 'inflamatorio'),
    (17, 'leve'),
    (18, 'contraindicar'),
    (19, 'precaucional'),
    (20, 'posológico'),
    (21, 'interaccionar'),
    (22, 'hipersensibilidad'),
    (23, 'sensibilización'),
    (24, 'sensibilizar'),
    (25, 'intolerante'),
    (26, 'nauseoso'),
    (27, 'somnoliento'),
    (28, 'irritativo'),
    (29, 'hemorrágico'),
    (30, 'hemorragia'),
    (31, 'discontinuar'),
    (32, 'retirar'),
    (33, 'pautar'),
    (34, 'terapéutica')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / Климат и экологические процессы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Климат и экологические процессы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Климат и экологические процессы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'climático'),
    (1, 'calentamiento'),
    (2, 'inundación'),
    (3, 'desertificación'),
    (4, 'ecosistema'),
    (5, 'hábitat'),
    (6, 'especie'),
    (7, 'carbono'),
    (8, 'metano'),
    (9, 'deforestación'),
    (10, 'acidificación'),
    (11, 'ambiental'),
    (12, 'regeneración'),
    (13, 'conservación'),
    (14, 'mitigar'),
    (15, 'descontaminar'),
    (16, 'descarbonizar'),
    (17, 'renaturalizar'),
    (18, 'fósil'),
    (19, 'hídrico'),
    (20, 'climatología'),
    (21, 'climatológico'),
    (22, 'calentarse'),
    (23, 'desertificar'),
    (24, 'desertificado'),
    (25, 'erosionar'),
    (26, 'erosivo'),
    (27, 'biodiverso'),
    (28, 'ecosistémico'),
    (29, 'invasor'),
    (30, 'contaminante'),
    (31, 'emisivo'),
    (32, 'deforestar'),
    (33, 'reforestar'),
    (34, 'acidificar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / Городское планирование
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Городское планирование'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Городское планирование'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'urbanismo'),
    (1, 'ciudad'),
    (2, 'zonificación'),
    (3, 'densidad'),
    (4, 'peatonalización'),
    (5, 'altura'),
    (6, 'edificabilidad'),
    (7, 'gentrificación'),
    (8, 'periferia'),
    (9, 'suburbio'),
    (10, 'metrópoli'),
    (11, 'conurbación'),
    (12, 'ordenamiento'),
    (13, 'mixto'),
    (14, 'compacto'),
    (15, 'disperso'),
    (16, 'caminable'),
    (17, 'urbanístico'),
    (18, 'ciclable'),
    (19, 'dotacional'),
    (20, 'densificar'),
    (21, 'recalificar'),
    (22, 'reurbanizar'),
    (23, 'zonificar'),
    (24, 'densificación'),
    (25, 'densificado'),
    (26, 'peatonalizar'),
    (27, 'reurbanización'),
    (28, 'reurbanizador'),
    (29, 'ordenanza'),
    (30, 'planeamiento'),
    (31, 'planeador'),
    (32, 'parcelación'),
    (33, 'parcelar'),
    (34, 'dotación')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / Транспорт, энергия, инфраструктура
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Транспорт, энергия, инфраструктура'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Транспорт, энергия, инфраструктура'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'ciclovía'),
    (1, 'subestación'),
    (2, 'transmisión'),
    (3, 'solar'),
    (4, 'eólico'),
    (5, 'hidroeléctrico'),
    (6, 'recarga'),
    (7, 'electrificación'),
    (8, 'intermodalidad'),
    (9, 'logística'),
    (10, 'túnel'),
    (11, 'viaducto'),
    (12, 'corredor'),
    (13, 'ferrocarril'),
    (14, 'eléctrico'),
    (15, 'electrificado'),
    (16, 'intermodal'),
    (17, 'logístico'),
    (18, 'multimodal'),
    (19, 'abastecimiento'),
    (20, 'tranviario'),
    (21, 'metropolitano'),
    (22, 'electrificar'),
    (23, 'electromovilidad'),
    (24, 'electromóvil'),
    (25, 'recargable'),
    (26, 'baterización'),
    (27, 'fotovoltaico'),
    (28, 'solarizado'),
    (29, 'multimodalidad'),
    (30, 'vial'),
    (31, 'tunelización'),
    (32, 'soterramiento'),
    (33, 'soterrar'),
    (34, 'abastecedor')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / География, регионы, ландшафты
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'География, регионы, ландшафты'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'География, регионы, ландшафты'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'región'),
    (1, 'territorio'),
    (2, 'llanura'),
    (3, 'meseta'),
    (4, 'costa'),
    (5, 'litoral'),
    (6, 'península'),
    (7, 'archipiélago'),
    (8, 'cuenca'),
    (9, 'desierto'),
    (10, 'selva'),
    (11, 'humedal'),
    (12, 'estepa'),
    (13, 'tundra'),
    (14, 'volcán'),
    (15, 'cordillera'),
    (16, 'altiplano'),
    (17, 'bahía'),
    (18, 'golfo'),
    (19, 'delta'),
    (20, 'cartografía'),
    (21, 'relieve'),
    (22, 'bioma'),
    (23, 'geográfico'),
    (24, 'paisajístico'),
    (25, 'costero'),
    (26, 'montañoso'),
    (27, 'fluvial'),
    (28, 'lacustre'),
    (29, 'desértico')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / Жильё, районы, качество среды
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Жильё, районы, качество среды'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Жильё, районы, качество среды'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'iluminación'),
    (1, 'proximidad'),
    (2, 'barrial'),
    (3, 'ruidoso'),
    (4, 'contaminado'),
    (5, 'ventilado'),
    (6, 'sombreado'),
    (7, 'próximo'),
    (8, 'equipado'),
    (9, 'deteriorado'),
    (10, 'rehabilitable'),
    (11, 'habitacional'),
    (12, 'residencialidad'),
    (13, 'vecinalidad'),
    (14, 'distrital'),
    (15, 'ambientalidad'),
    (16, 'salubridad'),
    (17, 'insalubre'),
    (18, 'acústica'),
    (19, 'sonoro'),
    (20, 'lumínico'),
    (21, 'ventilatorio'),
    (22, 'arboladura'),
    (23, 'arborización'),
    (24, 'arborizar'),
    (25, 'caminabilidad'),
    (26, 'deteriorar'),
    (27, 'gentrificar'),
    (28, 'insalubridad'),
    (29, 'insonorización')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- B2 / Экология, география, урбанистика / Катастрофы, устойчивость, ресурсы
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Катастрофы, устойчивость, ресурсы'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень B2'
    AND cat.name = 'Экология, география, урбанистика'
    AND ws.title = 'Катастрофы, устойчивость, ресурсы'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'catástrofe'),
    (1, 'desastre'),
    (2, 'terremoto'),
    (3, 'huracán'),
    (4, 'deslizamiento'),
    (5, 'avalancha'),
    (6, 'tsunami'),
    (7, 'alimento'),
    (8, 'reconstrucción'),
    (9, 'racionamiento'),
    (10, 'resistente'),
    (11, 'resiliente'),
    (12, 'agotable'),
    (13, 'catastrófico'),
    (14, 'desastroso'),
    (15, 'evacuar'),
    (16, 'evacuado'),
    (17, 'inundable'),
    (18, 'incendiario'),
    (19, 'sísmico'),
    (20, 'sismo'),
    (21, 'huracanado'),
    (22, 'tormentoso'),
    (23, 'deslizar'),
    (24, 'volcánico'),
    (25, 'eruptivo'),
    (26, 'colapsar'),
    (27, 'escaso'),
    (28, 'reconstruir'),
    (29, 'suministrar'),
    (30, 'racionar'),
    (31, 'abastecer'),
    (32, 'agotar'),
    (33, 'adaptativo'),
    (34, 'reservorio')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Академические связки и метатекст (1-30)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (1-30)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (1-30)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'igualmente'),
    (1, 'análogamente'),
    (2, 'simultáneamente'),
    (3, 'seguidamente'),
    (4, 'ulteriormente'),
    (5, 'consiguientemente'),
    (6, 'sucintamente'),
    (7, 'detalladamente'),
    (8, 'específicamente'),
    (9, 'particularmente'),
    (10, 'notablemente'),
    (11, 'metatexto'),
    (12, 'metatextual'),
    (13, 'introducción'),
    (14, 'epígrafe'),
    (15, 'inciso'),
    (16, 'ejemplificación'),
    (17, 'reformulación'),
    (18, 'expositivo'),
    (19, 'secuenciador'),
    (20, 'introductorio'),
    (21, 'aclaratorio'),
    (22, 'delimitador'),
    (23, 'transicional'),
    (24, 'digresivo'),
    (25, 'digresión'),
    (26, 'enumeración'),
    (27, 'enumerativo'),
    (28, 'tematización'),
    (29, 'tematizar')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Академические связки и метатекст (31-60)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (31-60)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (31-60)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'contrariamente'),
    (1, 'inversamente'),
    (2, 'recíprocamente'),
    (3, 'alternativamente'),
    (4, 'comparativamente'),
    (5, 'correlativamente'),
    (6, 'proporcionalmente'),
    (7, 'respectivamente'),
    (8, 'adicionalmente'),
    (9, 'complementariamente'),
    (10, 'subsidiariamente'),
    (11, 'preliminarmente'),
    (12, 'provisionalmente'),
    (13, 'tentativamente'),
    (14, 'presumiblemente'),
    (15, 'plausiblemente'),
    (16, 'verosímilmente'),
    (17, 'aparentemente'),
    (18, 'indudablemente'),
    (19, 'incuestionablemente'),
    (20, 'discutiblemente'),
    (21, 'cuestionablemente'),
    (22, 'paradójicamente'),
    (23, 'significativamente'),
    (24, 'sustancialmente'),
    (25, 'marginalmente'),
    (26, 'tangencialmente'),
    (27, 'incidentalmente'),
    (28, 'metodológicamente'),
    (29, 'conceptualmente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Академические связки и метатекст (61-70)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (61-70)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Академические связки и метатекст (61-70)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'terminológicamente'),
    (1, 'semánticamente'),
    (2, 'epistemológicamente'),
    (3, 'hermenéuticamente'),
    (4, 'discursivamente'),
    (5, 'pragmáticamente'),
    (6, 'normativamente'),
    (7, 'críticamente'),
    (8, 'ontológicamente'),
    (9, 'axiológicamente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Риторические маркеры и оттенки позиции (1-30)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (1-30)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (1-30)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'evidentemente'),
    (1, 'obviamente'),
    (2, 'naturalmente'),
    (3, 'indiscutiblemente'),
    (4, 'acaso'),
    (5, 'supuestamente'),
    (6, 'presuntamente'),
    (7, 'simplemente'),
    (8, 'meramente'),
    (9, 'solamente'),
    (10, 'únicamente'),
    (11, 'precisamente'),
    (12, 'justamente'),
    (13, 'relativamente'),
    (14, 'considerablemente'),
    (15, 'profundamente'),
    (16, 'radicalmente'),
    (17, 'moderadamente'),
    (18, 'cautelosamente'),
    (19, 'francamente'),
    (20, 'honestamente'),
    (21, 'ostensiblemente'),
    (22, 'palmariamente'),
    (23, 'manifiestamente'),
    (24, 'notoriamente'),
    (25, 'indudable'),
    (26, 'hipotéticamente'),
    (27, 'conjeturalmente'),
    (28, 'aproximativamente'),
    (29, 'sobradamente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Риторические маркеры и оттенки позиции (31-60)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (31-60)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (31-60)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'matizadamente'),
    (1, 'implícitamente'),
    (2, 'explícitamente'),
    (3, 'deliberadamente'),
    (4, 'intencionadamente'),
    (5, 'retóricamente'),
    (6, 'estratégicamente'),
    (7, 'enfáticamente'),
    (8, 'taxativamente'),
    (9, 'categóricamente'),
    (10, 'rotundamente'),
    (11, 'tajantemente'),
    (12, 'prudentemente'),
    (13, 'irónicamente'),
    (14, 'llamativamente'),
    (15, 'curiosamente'),
    (16, 'sorprendentemente'),
    (17, 'reveladoramente'),
    (18, 'sintomáticamente'),
    (19, 'elocuentemente'),
    (20, 'sugestivamente'),
    (21, 'problemáticamente'),
    (22, 'polémicamente'),
    (23, 'controvertidamente'),
    (24, 'razonablemente'),
    (25, 'justificadamente'),
    (26, 'legítimamente'),
    (27, 'sesgadamente'),
    (28, 'tendenciosamente'),
    (29, 'provocativamente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Риторические маркеры и оттенки позиции (61-70)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (61-70)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Риторические маркеры и оттенки позиции (61-70)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'aseverativamente'),
    (1, 'dubitativamente'),
    (2, 'incisivamente'),
    (3, 'contundentemente'),
    (4, 'impugnable'),
    (5, 'aseverativo'),
    (6, 'dubitativo'),
    (7, 'conjetural'),
    (8, 'incisivo'),
    (9, 'contundente')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Вежливость, имплицитность, дистанция (1-30)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Вежливость, имплицитность, дистанция (1-30)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Вежливость, имплицитность, дистанция (1-30)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'tacto'),
    (1, 'delicadeza'),
    (2, 'discreción'),
    (3, 'prudencia'),
    (4, 'cordialidad'),
    (5, 'insinuación'),
    (6, 'sugerencia'),
    (7, 'eufemismo'),
    (8, 'atenuación'),
    (9, 'rodeo'),
    (10, 'elipsis'),
    (11, 'silencio'),
    (12, 'sobreentendido'),
    (13, 'presuposición'),
    (14, 'insinuado'),
    (15, 'tácito'),
    (16, 'velado'),
    (17, 'deferente'),
    (18, 'discreto'),
    (19, 'indirecto'),
    (20, 'atenuado'),
    (21, 'mitigado'),
    (22, 'elíptico'),
    (23, 'deferencial'),
    (24, 'cortés'),
    (25, 'respetuosidad'),
    (26, 'reticencia'),
    (27, 'circunloquio'),
    (28, 'evasiva'),
    (29, 'sugestivo')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

-- C1 / Дискурс, регистр, прагматика / Вежливость, имплицитность, дистанция (31-60)
WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Вежливость, имплицитность, дистанция (31-60)'
  LIMIT 1
)
DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM target_set);

WITH target_set AS (
  SELECT ws.id
  FROM word_sets ws
  JOIN word_set_categories cat ON cat.id = ws.category_id
  JOIN word_set_categories parent ON parent.id = cat.parent_id
  WHERE ws.course_code = 'es_ru'
    AND parent.name = 'Уровень C1'
    AND cat.name = 'Дискурс, регистр, прагматика'
    AND ws.title = 'Вежливость, имплицитность, дистанция (31-60)'
  LIMIT 1
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'despersonalización'),
    (1, 'impersonalidad'),
    (2, 'deferencialidad'),
    (3, 'protocolariedad'),
    (4, 'formulismo'),
    (5, 'formalismo'),
    (6, 'cortesano'),
    (7, 'ceremonioso'),
    (8, 'ceremoniosidad'),
    (9, 'distanciado'),
    (10, 'reticente'),
    (11, 'circunloquial'),
    (12, 'evasivo'),
    (13, 'insinuativo'),
    (14, 'sobreentender'),
    (15, 'presuponer'),
    (16, 'tácitamente'),
    (17, 'veladamente'),
    (18, 'sutilmente'),
    (19, 'diplomáticamente'),
    (20, 'discretamente'),
    (21, 'eufemístico'),
    (22, 'elusivo'),
    (23, 'alusivo'),
    (24, 'circunspecto'),
    (25, 'comedido'),
    (26, 'mesurado'),
    (27, 'circunspección'),
    (28, 'comedimiento'),
    (29, 'mesura')
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;

