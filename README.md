# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

# Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

## Ejercicio N°1:
Se definió un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers sigue el formato propuesto: client1, client2, client3, etc. 

El script se ubica en la raíz del proyecto y recibe por parámetro el nombre del archivo de salida y la cantidad de clientes esperados.

### Cómo ejecutarlo

Se eligió usar la librería `pyyaml` debido a su facilidad de uso, compatibilidad, flexibilidad y simplicidad para manipular archivos YAML. Esto facilita la generación de archivos de configuración de Docker Compose de manera dinámica y personalizada.

Para instalar esta dependencia, se debe ejecutar el siguiente comando:

`pip install PyYAML`

Para generar el compose se debe ejecutar:

`./generar-compose.sh <archivo_salida> <num_clientes>`

Por ejemplo:

`./generar-compose.sh docker-compose-dev.yaml 5`

En caso de tener errores de permisos, se debe correr el siguiente comando para otorgarle a script permisos de ejecucion:

`chmod +x generar-compose.sh`

## Ejercicio N°2:
Se modificó el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) se inyecta en el container y es persistida por fuera de la imagen.

Para ello, se modificó el script `generar-compose.sh` incluyendo volúmenes que inyectan los archivos de configuración correspondientes en los contenedores.

* Servidor: `./server/config.ini:/config.ini`
* Cliente: `./client/config.yaml:/config.yaml`

## Ejercicio N°3:
Un script de bash `validar-echo-server.sh` fue creado en el cual se permite verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. 

El script se ubica en la raíz del proyecto. 

Para no exponer puertos del servidor para realizar la comunicación, el script crea un contenedor y se conecta al servidor mediante la red de docker. De esta manera, no expone el puerto por fuera de la red de contenedores. El contenedor creado es temporal y se ejecuta en la misma red del servidor: `tp0_testing_net`. 

Netcat no es instalado en la máquina _host_, sino que el script utiliza la imagen _alpine:latest_ para ejecutar `netcat` dentro de un contenedor temporal.

Dado que el servidor es un echo server, se envia un mensaje al servidor y espera recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa se imprime:

`action: test_echo_server | result: success`

En caso contrario, se imprimirá:

`action: test_echo_server | result: fail`.

### Cómo ejecutarlo

POr supuesto, la red debe estar corriendo, por lo que primero se debe correr:

`make docker-compose-up`

Luego, para correr el script:

`./validar-echo-server.sh`

En caso de tener errores de permisos, se debe correr el siguiente comando para otorgarle al script permisos de ejecución:

`chmod +x validar-echo-server.sh`

## Ejercicio N°4:
Se modificó el servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. 

### Servidor

El atributo `self.down` se utiliza para controlar el bucle principal del servidor y permitir que el servidor se detenga de manera ordenada cuando se recibe una señal SIGTERM. 

Además, se agrego un signal handler que captura la señal `signal.SIGTERM`. En caso que el servidor la reiba, se ejecutará la función establecida como handler `__handle_sigterm()`.

Es importante cerrar tanto el servidor como los sockets de cliente para asegurarse de que todos los recursos se liberen correctamente y evitar posibles fugas de recursos. Por lo tanto, esta función cierra tanto el socket del cliente como el suyo para no aceptar más conexiones. Por último, el atributo `self.down` se setea en _True_, para que el bucle principal termine.

### Cliente

En el caso del cliente, para manejar el apagado ordenado (graceful shutdown) al recibir la señal SIGTERM, uso el paquete _os/signal_ en Go para capturar la señal y cerrar los recursos de manera adecuada.

El cliente maneja la señal SIGTERM mediante la función `HandleSigterm`. Esta función crea un canal para recibir señales y utiliza `signal.Notify` para capturar la señal. Cuando se recibe la señal, se llama a la función `Shutdown` para cerrar los recursos.

### Cómo ejecutarlo

Para probar que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM se deben correr los siguientes comandos.

1. Iniciar los contenedores

`make docker-compose-up`

2. Enviar la señal SIGTERM

`docker stop server`
`docker stop client1`

3. Verificar los logs

`make docker-compose-logs`

Para verificar el correcto funcionamiento, se loguearon mensajes en el cierre de cada recurso. Los mensajes que se pueden observar son:

```
server   | 2025-03-24 12:13:19 INFO     action: sigterm_received | result: in_progress
server   | 2025-03-24 12:13:19 INFO     action: close_client_socket | result: in_progress
server   | 2025-03-24 12:13:19 INFO     action: close_client_socket | result: success
server   | 2025-03-24 12:13:19 INFO     action: close_server_socket | result: in_progress
server   | 2025-03-24 12:13:19 INFO     action: close_server_socket | result: success
server   | 2025-03-24 12:13:19 INFO     action: sigterm_received | result: success
```

```
client1  | 2025-03-24 12:13:15 INFO     action: sigterm_received | result: in_progress | signal: terminated
client1  | 2025-03-24 12:13:15 INFO     action: shutdown | result: in_progress | client_id: 1
client1  | 2025-03-24 12:13:15 INFO     action: close_connection | result: in_progress | client_id: 1
client1  | 2025-03-24 12:13:15 INFO     action: close_connection | result: success | client_id: 1
client1  | 2025-03-24 12:13:15 INFO     action: shutdown | result: success | client_id: 1
client1  | 2025-03-24 12:13:15 INFO     action: sigterm_received | result: success | signal: terminated
client1  | 2025-03-24 12:13:16 INFO     action: loop_finished | result: success | client_id: 1
```

# Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. 

## Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

### Cliente
Emula a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

La lógica del cliente consiste en:
1. Leer datos de la apuesta desde variable de entorno.
2. Envío de la apuesta al servidor mediante el protocolo correspondiente.
3. Recepción de confirmación del servidor
4. Se imprimen logs.

### Servidor
Emula a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacena la información mediante la función `store_bet(...)` para control futuro de ganadores.

La lógica del servidor consiste en:
1. Recepción de apuesta mediante el protocolo correspondiente.
2. Almacenamiento de la misma mediante la función `store_bet(...)`.
3. Envío de confirmación al cliente.
4. Se imprimen logs.

### Protocolo de comunicación
El protocolo de comunicación entre el cliente y el servidor se basa en el envío y recepción de paquetes estructurados. Cada mensaje consta de un encabezado (header) y un cuerpo (body). A continuación se describe la estructura de cada parte del mensaje:

#### Análisis de los datasets
Para ser más precisa en el tamaño de cada campo, realicé un análisis de los datasets. Para los campos de _Documento_ y _Fecha de nacimiento_ no había problema, porque no varían la longitud. Sin embargo, los campos _Nombre_ y _Apellido_ pueden tener mucha variación en la cantidad de caracteres. Respecto a estos dos campos, encontré qué la máxima longitud de un nombre es de 23 caracteres mientras que la máxima para un apellido es de 10 caracteres.

Corriendo un script de Python obtuve los siguientes resultados:
```
Nombre más largo: Milagros De Los Angeles
Apellido más largo: Demichelis
Número más grande: 9999
```

Este análisis se realizó para disminuir la cantidad de bytes por cada apuesta, con el objetivo de maximizar la cantidad de apuestas por paquete.

#### Encabezado (header)

El encabezado del mensaje contiene la longitud del mensaje y el ID de la agencia. 

La estructura es la siguiente:
* Longitud del mensaje: 2 bytes (uint16) que indican la longitud total del mensaje, incluyendo el encabezado y el cuerpo.
* ID de la agencia: 1 byte (uint8) que indica el ID de la agencia que envía el mensaje.

#### Cuerpo (body)

El cuerpo del mensaje contiene la información de la apuesta. 

La estructura es la siguiente:

1. Longitud del nombre: 1 byte (uint8) que indica la longitud del nombre. Esto es suficiente ya que el máximo es 23. De todas maneras, se permiten nombres más largos hasta los 255 caracteres.
2. Nombre: N bytes, donde N es la longitud del nombre.
3. Longitud del apellido: 1 byte (uint8) que indica la longitud del apellido. Esto es suficiente ya que el máximo es 10. De todas maneras, se permiten apellidos más largos hasta los 255 caracteres.
4. Apellido: M bytes, donde M es la longitud del apellido.
5. Documento: 4 bytes (uint32) que representan el documento de la persona. Esto es suficiente ya que todos los documentos tienen 8 caracteres.
6. Fecha de nacimiento: 10 bytes (cadena de caracteres) que representan la fecha de nacimiento en formato _YYYY-MM-DD_.
7. Número de apuesta: 2 bytes (uint16) que representan el número de apuesta. Si bien no superan los 4 dígitos, se permiten números hasta 65536.

#### Ejemplo

Supongamos que tenemos la siguiente apuesta: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574`. Además, supongamos que el cliente con `client_id=1` le envía el mensaje al servidor.

La estructura del mensaje sería la siguiente:

1. Encabezado:
    * Longitud del mensaje: 51 bytes (2 bytes)
    * ID de la agencia: 1 (1 byte)

2. Cuerpo
    * Longitud del nombre: 15 (1 byte)
    * Nombre: "Santiago Lionel" (15 bytes)
    * Longitud del apellido: 5 (1 byte)
    * Apellido: "Lorca" (5 bytes)
    * Documento: 30904465 (4 bytes)
    * Fecha de nacimiento: "1999-03-17" (10 bytes)
    * Número: 7574 (2 bytes)

#### Respuesta del servidor

Por otro lado, el servidor tiene 2 códigos de respuesta:

1. **OK (0)**: la apuesta se almacenó correctamente.
2. **ERROR (1)**: no se pudo almacenar la apuesta debido a algún error.

#### Lectura de los mensajes

Del lado del cliente, esto es más simple ya que el servidor solo responde un byte. De esta manera, la forma de leer del lado del cliente es simplemente leer un byte.

Por el otro lado, el servidor lee primero dos bytes (definido por `MSG_SIZE_LEN`). A partir de esto, sabe la longitud del mensaje. Por lo tanto, lee del socket la longitud del mensaje correspondiente.

Luego, para deserializar el mensaje:
* El primer byte es el agency_id definido por `AGENCY_ID_LEN`.
* El resto del mensaje corresponde a la apuesta que se deserializa teniendo en cuenta la estructura descrita anteriormente.

### Cómo ejecutarlo

1. Iniciar los contenedores

`make docker-compose-up`

2. Verificar los logs

`make docker-compose-logs`

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

Por lo tanto, para el mismo ejemplo anterior, se deberán observar los siguientes logs:

```
server   | 2025-03-24 13:02:59 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2025-03-24 13:02:59 INFO     action: receive_message | result: success | agency_id: 1
server   | 2025-03-24 13:02:59 INFO     action: apuesta_almacenada | result: success | dni: 30904465 | numero: 7574
```

En los logs del server se puede observar que recibe el mensaje de la agencia, deserializa la apuesta y la almacena. Finalmente, se imprime el log de `apuesta_almacenada`.

```
client1  | 2025-03-24 13:02:59 INFO     action: send_message | result: in_progress | client_id: 1
client1  | 2025-03-24 13:02:59 INFO     action: receive_message | result: success | client_id: 1 | response_status: 0
client1  | 2025-03-24 13:02:59 INFO     action: apuesta_enviada | result: success | dni: 30904465 | numero: 7574
```

Del lado del cliente, se puede observar que envía el mensaje. Luego, recibe respuesta del servidor. La respuesta es un 0, por lo que significa que el servidor respondió un OK, ya que lo pudo almacenar. Finalmente, se imprime el log de `apuesta_enviada`.

## Ejercicio N°6:
Se realizó una modificación en los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

### Cliente
Se modificó el cliente para leer las apuestas de un archivo CSV, en vez de recibir una apuesta por variable de entorno. Para ello, el cliente utiliza un lector CSV (`csv.Reader`) para leer las apuestas del archivo en batchs. La cantidad máxima de apuestas en cada batch se configura en `config.yaml` bajo la clave `batch: maxAmount`.

#### Configuración de Volúmenes para Archivos CSV

Para permitir que los clientes accedan a los archivos CSV, se configuraron volúmenes en el archivo `docker-compose-dev.yaml` generado por el script `generar-compose.sh`. Estos volúmenes montan el directorio local `.data` en el contenedor del cliente, permitiendo que los archivos CSV sean accesibles desde dentro del contenedor.

Para ello, es importante contar con un archivo con la siguiente ruta y estructura:

`/.data/agency-<client_id>.csv`

### Servidor
La lógica del servidor continúa siendo la misma, con la única diferencia que en un mismo mensaje debe deserializar más de una apuesta, para posteriormente almacenar cada una.

### Protocolo de comunicación
Se realizaron algunas modificaciones al protocolo. 

El cliente puede enviar dos tipos de mensajes:

1. **BATCH (0)**: el mensaje enviado contiene una cantidad de apuestas dadas por `batch: maxAmount`.
2. **BATCH_END (1)**: este mensaje es usado por el cliente para indicar que finalizó con el envio de apuestas.

#### Encabezado (header)

De acuerdo a que ahora hay distintos tipos de mensaje, se le agregó al encabezado del mensaje el tipo de mensaje. 

La estructura es la siguiente:
* Longitud del mensaje: 2 bytes (uint16) que indican la longitud total del mensaje, incluyendo el encabezado y el cuerpo.
* ID de la agencia: 1 byte (uint8) que indica el ID de la agencia que envía el mensaje.
* Tipo de mensaje: 1 byte (uint8) que indica qué tipo de mensaje se está enviando.

#### Cuerpo (body)

Se mantiene igual que antes, con la diferencia que ahora puede haber más de una apuesta.

De acuerdo a lo explicado previamente en la estructura del cuerpo, una apuesta ocupa una .longitud de 51 bytes donde:

```
1 byte de longitud de nombre + 23 bytes de nombre + 1 byte de longitud de apellido + 10 bytes de apellido + 4 bytes documento + 10 bytes fecha de nacimiento + 2 bytes número = 51
```

La cantidad máxima de apuestas dentro de cada _batch_ es configurable desde _config.yaml_. Como los paquetes no deben exceder los 8kB, si una apuesta ocupa 51 bytes, la máxima cantidad de apuestas que se pueden enviar son 156. Esto equivale a 7956 bytes. Luego, se define la clave `batch: maxAmount` con un valor de 156. Por otro lado, el header del paquete ocupa 4 bytes.

#### Respuesta del servidor

La respuesta del servidor sigue siendo la misma. El servidor responde un **OK** cuando termina de procesar y almacenar cada apuesta en el batch. Esto es útil para no saturar al servidor, ya que el cliente espera por esta confirmación antes de enviar el siguiente mensaje.
Por supuesto, el servidor responde con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

#### Lectura de los mensajes

La lectura de los mensajes es igual que antes, con la diferencia que se agregó el campo de tipo de mensaje.

El servidor lee primero dos bytes (definido por `MSG_SIZE_LEN`). A partir de esto, sabe la longitud del mensaje. Por lo tanto, lee del socket la longitud del mensaje correspondiente.

Luego, para deserializar el mensaje:
* El primer byte es el agency_id definido por `AGENCY_ID_LEN`.
* El segundo byte es el message_type definido por `MSG_TYPE_LEN`.
* El resto del mensaje corresponde a las apuestas que se deserializan teniendo en cuenta la estructura descrita anteriormente.

### Cómo ejecutarlo

1. Iniciar los contenedores

`make docker-compose-up`

2. Verificar los logs

`make docker-compose-logs`

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

Por ejemplo, se observan los siguientes logs:

```
client1  | 2025-03-24 14:50:36 INFO     action: send_batch_message | result: success| client_id: 1 | batch_length: 156
server   | 2025-03-24 14:50:41 INFO     action: apuesta_recibida | result: success | cantidad: 156
```

Al finalizar de enviar todo se observa:

```
client1  | 2025-03-24 14:50:56 INFO     action: send_end_message | result: success | client_id: 1
server   | 2025-03-24 14:50:56 INFO     action: client_finished_sending_bets | result: success | agency: 1
```
