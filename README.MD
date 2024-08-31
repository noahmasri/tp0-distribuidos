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

## Ejercicio 3
Para llevar a cabo lo pedido, se creó un script que crea un contenedor que envía un mensaje al servidor ("Hello echoer!") y se conecta a este mediante la red de docker para no deber exponer el puerto del servidor por fuera de la red de contenedores. Este contenedor creado es temporal, es decir, luego de terminar su ejecución es eliminado, y se conecta a la red que define el archivo `docker-compose-dev.yaml`, tp0_testing_net.

Este script se basa principalmente en esta porción de código, que es la que se encarga de enviar el mensaje al servidor, y de obtener su respuesta. Esto lo que hace es guardarse en una variable response únicamente el resultado de la ejecución del comando `netcat`, deshaciendose tanto de la salida de stdout como la de stderr de todos los comandos. Lo primero que hace es instalar `netcat`, para lo cual había 2 versiones, traditional y openbsd, por lo que se debió elegir una. Luego hace echo del mensaje a enviar, enviando esto al servidor en el puerto en que este se encuentra, estableciendo arbitrariamente un timeout de 1 segundo.

```bash
RESPONSE=$(docker run --rm --network tp0_testing_net --name netcat ubuntu:18.04 bash -c "
    apt-get update > /dev/null 2>&1 && \
    apt-get install -y netcat-traditional > /dev/null 2>&1 && \ 
    echo \"$TEST_MESSAGE\" | nc -w 1 "server" $PORT
" 2>&1)
```
En la variable response entonces se guarda únicamente la respuesta del servidor. Se redirige a la salida estándar la salida de error para posible debugging. De querer saber por qué está fallando, se puede hacer echo del mensaje de response.

### Como ejecutarlo
Para que esto funcione, se debe tener la red corriendo, y para esto debemos correr el comando:
```bash
make docker-compose-up
```
En caso de que esta no esté corriendo, se recibirá el output `action: test_echo_server | result: fail`.
Una vez hecho esto correr:
```bash
./validar-echo-server.sh
```
En caso de recibir el error `bash: ./validar-echo-server.sh: Permission denied`, corra `chmod +x validar-echo-server.sh` para otorgar permisos de ejecución al script, y luego vuelva a ejecutar el comando anterior.
Si todo se encuentra corriendo y se recibe por parte del servidor el mismo mensaje que se le envió, se espera que el output sea `action: test_echo_server | result: success`.

## Ejercicio 4
Para este ejercicio se pidió que ambos el servidor y el cliente, al recibir una señal SIGTERM realicen un graceful shutdown. Para llevar a cabo esto en el servidor, se agregó a este un signal handler, que funciona mediante eventos. Si el servidor recibe la señal, automáticamente va a comenzar a ejecutarse la función establecida como handler, en este caso el método `__shutdown_gracefully`. Esta función lo que hace es settear el atributo `should_stop` en true, y cerrar el socket para no aceptar más conexiones. Al regresar a donde estaba, probablemente reciba un error intentando acceder al socket, por lo que este atributo sirve para ignorar el error que podría darse, ya que no es un error del SO sino que fue cerrado apropósito. A su vez, hice que el servidor se guarde como atributo el socket del cliente con el que se está comunicando para también cerrar este.
Por otro lado, la forma de manejar el shutdown desde el cliente es levemente distinta, debido a como maneja go las excepciones. Para manejar las excepciones en este, se debe escuchar mediante un channel, que en mi caso se llama sigchan. Además de esto, se agregó un channel para comunicar el hilo principal con la subrutina que ejecuta el bucle del cliente, mediante el cual el cliente envía un booleano por el canal para avisar que terminó cuando termina el loop, o cuando ocurre un error en la lectura del socket. La función main tiene dos formas de terminar: si recibe un done, donde simplemente retorna, o si recibe la señal por el sigchan, donde ejecuta la función `ShutdownGracefully()`.

### Como ejecutarlo
Para verificar el correcto manejo de las señales, se debe correr:
```bash
make docker-compose-up
```
para iniciar los contenedores, y luego ejecutar:
```bash
make docker-compose-stop
```
para enviar la señal a los contenedores sin eliminarlos. Una vez hecho esto, se pueden observar los logs corriendo:
```bash
make docker-compose-logs
```
para ver las distintas acciones que ejecuta cada proceso al recibir SIGTERM. 

En caso de querer ver la ejecución llegar a su fin, saltear el segundo paso y acceder a los logs. Una vez visto lo deseado en los logs, puede salir de estos con ctrl+c, y cortar la ejecución de los contenedores con:
```bash
make docker-compose-down
```

# Parte 4
## Ejercicio 5
Para este ejercicio se pidió crear un protocolo de comunicación que permita que se emule una agencia de lotería, enviando por cliente una única apuesta. Este protocolo debe contener información suficiente para que se puedan evitar los short writes y los short reads.
Antes de desarrollar el protocolo en sí, se realizó un pequeño análisis de los datos otorgados por la cátedra, donde se encontró que:
* La máxima longitud de un nombre es de 23 caracteres en todos los csvs, mientras que la mínima se encuentra en la agencia 2, y es de 6 caracteres. Además, que una porción muy poco significativa de entradas tienen un apellido con ese máximo (menos de un 0,2%)
* La máxima longitud de un apellido es de 10 caracteres, mientras que la mínima se encuentra en las agencias 1 y 3 y es de 6 caracteres. Además, que una porción pequeña de entradas tienen un apellido con ese máximo (menos de un 2%)

Y además, se corroboró que:
* Todos los documentos tienen 8 caracteres. Por como están estructurados, sabemos que el máximo valor que podrían tomar es 99.999.999 (aunque en los datasets sea 40000000 aprox), valor el cual, eliminando los '.', entra en un u32.
* Las fechas tienen formato "YYYY-MM-DD"
* Las apuestas son todas de 4 digitos, numéricas, y con un valor máximo de 9999

Teniendo en cuenta todas estas validaciones sobre los datos es que se desarrolló un protocolo que pretende maximizar la cantidad de apuestas que se pueden realizar por paquete o por segundo, disminuyendo la cantidad de bytes utilizados por cada una de estas. El protocolo entonces plantea que en un mensaje viajará en el siguiente orden la siguiente información:

* Agencia: 1 byte, ya que por enunciado hay únicamente 5. 
* Longitud del nombre: para esto, dado que el máximo es 23 caracteres, se asigna 1 byte, u8, que será suficiente para alojar el tamaño. Se permite que puedan aparecer nombres más largos, siempre y cuando no superen los 255 caracteres
* Nombre: cantidad de bytes que se corresponden con la longitud. En base al byte anterior, se sabrá cuántos se deberá leer para obtener el nombre entero. 
* Longitud del apellido: para esto, dado que el máximo es 10 caracteres, se asigna 1 byte. Se permite que puedan aparecer nombres más largos, siempre y cuando no superen los 255 caracteres.
* Apellido: cantidad de bytes que se corresponden con la longitud.
* Documento: campo de longitud fija de 4 bytes, un u32, que permite alojar números que llegan hasta 4294967295, por lo que 100 millones entran en este tipo de números.
* Fecha: campo de longitud fija de 10 bytes. Se incluyen los '-', ya que es práctico almacenarlos así, y se considera que se podría tardar más parseando y desparseando este campo.
* Numero de apuesta: campo de longitud fija de 2 bytes, un u16, que permite alojar números que llegan hasta 65536, por lo que cualquier numero hasta el 9999 entran en este tipo de números.

Finalmente, se definen para el servidor 3 códigos de respuesta:
* Ok (0): si se pudo almacenar toda la información correctamente
* Error (1): si no se pudo almacenar la información, o no se la pudo leer por completo
* Abort (2): si el servidor recibió un sigterm o debió finalizar abruptamente por alguna señal

### Como ejecutarlo
Antes de ejecutarlo, debemos asegurarnos de tener las variables de entorno de la apuesta correctamente configuradas. Para esto, se debe encontrar en el cliente, dentro de la seccion environment del docker-compose.yaml algo similar a:
```yaml
environment:
    - CLI_ID=2
    - CLI_LOG_LEVEL=DEBUG
    - BET_NAME=<name of betmaker>
    - BET_SURNAME=<name of betmaker>
    - BET_ID=<Id number of betmaker>
    - BET_BIRTHDATE=<birthdate of betmaker in iso format (yyyy-mm-dd)>
    - BET_NUMBER=<number bet corresponds to>
```
Tal como en los anteriores ejercicios, debemos ejecutarlo corriendo los distintos targets del makefile. En este caso correríamos, para iniciar el servidor y los clientes:
```bash
make docker-compose-up
```
Para ver los logs de las apuestas:
```bash
make docker-compose-logs
```
Para finalizar la ejecución de todos los contenedores:
```bash
make docker-compose-down
```

## Ejercicio 6
Para este ejercicio, se pide leer apuestas de un csv, y enviarlas de a tandas al servidor. Para este, en primer lugar se pide definir la máxima cantidad de apuestas que se enviarán en un lote, si este lote no puede superar los 8kB de data. Dado el protocolo definido, sabemos que se envian por cada apuesta:
* long nombre: 1 byte por apuesta
* nombre: hasta 255 bytes por apuesta
* long apellido: 1 byte por apuesta
* apellido: hasta 255 bytes por apuesta
* documento: 4 bytes por apuesta
* fecha: 10 bytes por apuesta
* numero apuesta: 2 bytes por apuesta
Lo cual nos deja un máximo de 528 bytes por apuesta. En base a la investigación previa del set de datos, sabemos que el máximo en realidad es con 23 caracteres para el nombre, y 10 para el apellido, lo que nos da entonces en máximo de 51 bytes por apuesta. Con este número, tenemos entonces que la máxima cantidad de apuestas que entran en un batch son 156, ocupando 7956 bytes en el peor de los casos, y dejando la cantidad suficiente de bytes para agregar el número de agencia, y un campo nuevo que indique la cantidad de apuestas en el batch. 

# Referencias
[Instalar netcat en contenedor](https://stackoverflow.com/questions/52570028/i-am-unable-to-execute-netcat-command-within-docker-bash-terminal)

[Solucionar problema: 'netcat' has no installation candidate](https://forums.docker.com/t/package-netcat-has-no-installation-candidate-how-to-fix-this/136541)

[Correr comandos de terminal en docker container](https://www.warp.dev/terminus/docker-run-bash)

[Librería para el manejo de señales en go](https://go.dev/pkg/os/signal/?m=old)

[Ejemplos de manejo de señales en go](https://gobyexample.com/signals)

[Ejemplos de waitgroups en go](https://gobyexample.com/waitgroups)
