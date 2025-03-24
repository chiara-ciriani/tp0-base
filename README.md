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
