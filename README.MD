# Parte 1

## Ejercicio 1
En el archivo `docker-compose-dev.yaml` se encuentra el archivo compose original, con el cual se realizaron pruebas para ver si la estructura que iba a desarrollar era la apropiada. Luego, en el archivo `docker-compose-test.yaml` se encuentra un archivo generado con el script de bash. Si bien la estructura de estos difiere levemente por problemas de indentación con la librería usada, su ejecución funciona como fue pedido. 

Se eligió usar la librería `pyyaml` en lugar de usar strings con el formato hecho a mano ya que considero que es mucho más seguro ante errores humanos. El uso de diccionarios para establecer las estructuras que luego mediante la propia librería se genera la indentación, hace que el código sea mucho mas legible y por tanto, menos propenso a errores.

### Como ejecutarlo
Antes de ejecutar el script, deberán asegurarse de contar con la dependencia requerida, en este caso, pyyaml. Para esto, se debe ejecutar el comando:
```bash
pip install PyYAML
```

Para generar el compose, debe correr el comando:
```bash
./generar-compose.sh <nombre archivo output> <numero de clientes>
```
En caso de recibir el error `bash: ./generar-compose.sh: Permission denied`, corra `chmod +x generar-compose.sh` para otorgar permisos de ejecución al script.

Es recomendable correr el comando con **docker-compose-dev.yaml** como nombre del archivo de output, ya que el makefile corre este archivo. En caso de desear otro nombre, deberá modificar el makefile. Una vez generado el compose, para ejecutar el programa, deberá usar los targets del makefile previamente mencionado. Los targets se ejecutan mediante la invocación de:

* **make \<target\>**:
Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de _debugging_ y _troubleshooting_.

Los targets disponibles son:
* **docker-compose-up**: Inicializa el ambiente de desarrollo (buildear docker images del servidor y cliente, inicializar la red a utilizar por docker, etc.) y arranca los containers de las aplicaciones que componen el proyecto.
* **docker-compose-down**: Realiza un `docker-compose stop` para detener los containers asociados al compose y luego realiza un `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene.
* **docker-compose-logs**: Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose.
* **docker-image**: Buildea las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para testear nuevos cambios en las imágenes antes de arrancar el proyecto.
* **build**: Compila la aplicación cliente para ejecución en el _host_ en lugar de en docker. La compilación de esta forma es mucho más rápida pero requiere tener el entorno de Golang instalado en la máquina _host_.

## Ejercicio 2
Se genera el ejercicio 2 a partir de lo realizado para el ejercicio 1. En esta rama se encuentran modificados tanto los dos compose que se encontraban de antes, como el generator para agregar un volume a cada contenedor que mapee de su archivo de configuración externo (ubicado en el sistema de archivos del host) al interior del contenedor.

Para verificar que el archivo de configuración se encuentre correctamente montado se puede usar el comando:

```bash
docker inspect <nombre_del_contenedor> --format='{{json .Mounts}}'
```
donde se espera, por ejemplo para el servidor, el output:
```bash
[{"Type":"bind","Source":"$(CWD)/server/config.ini","Destination":"/config.ini","Mode":"rw","RW":true,"Propagation":"rprivate"}]
```
siendo CWD el directorio en donde se encuentra el archivo compose.

Para ejecutar el EchoServer siga los mismos pasos que para el **ej1**.