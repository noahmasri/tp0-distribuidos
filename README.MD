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
Por otro lado, la forma de manejar el shutdown desde el cliente es levemente distinta, debido a como maneja go las excepciones. Para manejar las excepciones en este, se debe escuchar mediante un channel, que en mi caso se llama notifier o sigchan. Este canal es escuchado en una goroutine que se encarga de, o bien manejar la señal recibida, o bien salir cuando se cierra el canal de comunicación con el hilo principal. Dentro del hilo principal, al terminar su ejecución se cierra el canal "done" para que la goroutine que maneja la excepción sepa que terminó y salga. Para que el hilo principal, se usa una barrera (waitgroup en go) para que este espere a la goroutina. Para que el hilo principal se entere si saltó la señal, se usa el mismo channel done, y en cada iteración este se fija si terminó el timer o recibió algo por done para ver si iterar una vez más o finalizar.

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

# Referencias
[Instalar netcat en contenedor](https://stackoverflow.com/questions/52570028/i-am-unable-to-execute-netcat-command-within-docker-bash-terminal)

[Solucionar problema: 'netcat' has no installation candidate](https://forums.docker.com/t/package-netcat-has-no-installation-candidate-how-to-fix-this/136541)

[Correr comandos de terminal en docker container](https://www.warp.dev/terminus/docker-run-bash)

[Librería para el manejo de señales en go](https://go.dev/pkg/os/signal/?m=old)

[Ejemplos de manejo de señales en go](https://gobyexample.com/signals)
