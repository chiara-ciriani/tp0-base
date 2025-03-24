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
