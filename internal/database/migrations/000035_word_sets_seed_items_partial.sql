-- Seed es_ru word-set items for the currently filled portion of resources/wordsets/es_ru_seed_word_sets.json
-- (166/421 sets, 3000/10000 words at generation time). Minimal word_cards only;
-- TrainingWorker fills definitions/training cards/verb links and PronunciationService
-- schedules TTS for any new word asynchronously after this migration runs.
-- Remaining sets stay empty until resources/wordsets/es_ru_seed_word_sets.json is completed
-- and a follow-up migration seeds the rest.

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
  ('esqueleto', '', 'es_ru', CURRENT_TIMESTAMP)
ON CONFLICT(word) DO UPDATE SET
  course_code = COALESCE(word_cards.course_code, 'es_ru'),
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'este'),
    (1, 'ese'),
    (2, 'aquel'),
    (3, 'mi'),
    (4, 'tu'),
    (5, 'su')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'sí'),
    (1, 'no'),
    (2, 'muy'),
    (3, 'también'),
    (4, 'tampoco')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'país'),
    (1, 'idioma'),
    (2, 'nacionalidad'),
    (3, 'español'),
    (4, 'extranjero')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'persona'),
    (1, 'amigo'),
    (2, 'hombre'),
    (3, 'mujer')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'querer'),
    (1, 'poder'),
    (2, 'saber'),
    (3, 'entender'),
    (4, 'necesitar'),
    (5, 'deber'),
    (6, 'comprender')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'bolso'),
    (1, 'mochila'),
    (2, 'llave'),
    (3, 'dinero'),
    (4, 'pasaporte'),
    (5, 'documento'),
    (6, 'tarjeta')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'teléfono'),
    (1, 'escuela'),
    (2, 'clase'),
    (3, 'internet'),
    (4, 'mensaje')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'calle'),
    (1, 'plaza'),
    (2, 'tienda'),
    (3, 'banco'),
    (4, 'hotel'),
    (5, 'restaurante')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'autobús'),
    (1, 'tren'),
    (2, 'taxi'),
    (3, 'billete'),
    (4, 'comprar'),
    (5, 'pagar')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'día'),
    (1, 'mañana'),
    (2, 'tarde'),
    (3, 'noche'),
    (4, 'ahora')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'carné'),
    (1, 'permiso'),
    (2, 'visado'),
    (3, 'certificado'),
    (4, 'copia')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'alojamiento'),
    (1, 'recepción'),
    (2, 'huésped'),
    (3, 'alquilar'),
    (4, 'alquiler'),
    (5, 'llavero')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'maleta'),
    (1, 'equipaje'),
    (2, 'viaje'),
    (3, 'ida'),
    (4, 'vuelta'),
    (5, 'pasaje'),
    (6, 'destino')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), seed_words(sort_order, lemma) AS (
  VALUES
    (0, 'veinte'),
    (1, 'treinta'),
    (2, 'cuarenta'),
    (3, 'cincuenta'),
    (4, 'sesenta'),
    (5, 'setenta'),
    (6, 'último')
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
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
), cleared AS (
  DELETE FROM word_set_items
  WHERE word_set_id IN (SELECT id FROM target_set)
)
INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
SELECT target_set.id, wc.id, seed_words.sort_order
FROM target_set
JOIN seed_words ON TRUE
JOIN word_cards wc ON LOWER(wc.word) = LOWER(seed_words.lemma)
ORDER BY seed_words.sort_order;
