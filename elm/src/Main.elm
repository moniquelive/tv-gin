module Main exposing (..)

import Array exposing (Array)
import Browser
import Dict exposing (Dict)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Url exposing (percentEncode)



-- MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- MODEL


type alias Meme =
    { id : String
    , name : String
    , filename : String
    , font_color : String
    , line_chars : Int
    , boxes : List (List Int)
    }


defaultMeme : Meme
defaultMeme =
    Meme "" "" "" "" 0 []


type State
    = Failure
    | Loading
    | Success


type alias Model =
    { memes : Dict String Meme
    , selectedMeme : String
    , texts : Array String
    , tempTexts : Array String
    , state : State
    }


defaultModel =
    Model Dict.empty "" Array.empty Array.empty Loading


init : () -> ( Model, Cmd Msg )
init _ =
    ( defaultModel, getConfig )



-- UPDATE


type Msg
    = RefreshConfig
    | GotConfig (Result Http.Error (Dict String Meme))
    | MemeChanged String
    | TextChanged Int String
    | Generate


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RefreshConfig ->
            ( defaultModel, getConfig )

        GotConfig (Ok memes) ->
            let
                firstMeme =
                    memes |> Dict.values |> List.head |> Maybe.withDefault defaultMeme

                numTexts =
                    List.length firstMeme.boxes

                texts =
                    Array.repeat numTexts ""

                tempTexts =
                    Array.repeat numTexts ""
            in
            ( Model memes firstMeme.id texts tempTexts Success, Cmd.none )

        GotConfig (Err _) ->
            ( { model | state = Failure }, Cmd.none )

        MemeChanged id ->
            let
                meme =
                    Dict.get id model.memes |> Maybe.withDefault defaultMeme

                numTexts =
                    List.length meme.boxes

                nthTempText n =
                    Array.get n model.tempTexts |> Maybe.withDefault ""

                tempTexts =
                    Array.initialize numTexts nthTempText

                nthText n =
                    Array.get n model.texts |> Maybe.withDefault ""

                texts =
                    Array.initialize numTexts nthText
            in
            ( { model | selectedMeme = id, texts = texts, tempTexts = tempTexts }, Cmd.none )

        TextChanged index str ->
            ( { model | tempTexts = Array.set index str model.tempTexts }, Cmd.none )

        Generate ->
            ( { model | texts = model.tempTexts }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.state of
        Failure ->
            div []
                [ text "Não consegui carregar a configuração inicial. Tente novamente. "
                , button [ class "btn btn-primary", onClick RefreshConfig ] [ text "De novo!" ]
                ]

        Loading ->
            text "Carregando..."

        Success ->
            let
                allMemes =
                    Dict.values model.memes |> List.map viewMeme

                allText =
                    Array.toIndexedList model.tempTexts |> List.map createText
            in
            div []
                [ h1 [] [ text "Escolha o Meme:" ]
                , div [ class "row" ]
                    [ div [ class "col-md-4 mb-4 mb-md-0" ]
                        [ select [ value model.selectedMeme, class "mb-4 form-control", onInput MemeChanged ]
                            [ optgroup [ attribute "label" "Disponíveis" ] allMemes ]
                        , fieldset [] allText
                        , button [ class "mt-4 btn btn-primary", onClick Generate ] [ text "Gerar!" ]
                        ]
                    , div [ class "col-md-8" ]
                        [ img [ class "mw-100", src ("/meme?meme=" ++ model.selectedMeme ++ "&" ++ encodeTexts model.texts) ]
                            []
                        ]
                    ]
                ]


viewMeme : Meme -> Html Msg
viewMeme meme =
    option [ value meme.id ] [ text meme.name ]


createText : ( Int, String ) -> Html Msg
createText ( index, str ) =
    let
        isEnter code =
            if code == 13 then
                D.succeed Generate

            else
                D.fail "not ENTER"
    in
    input
        [ onInput (TextChanged index), on "keydown" (D.andThen isEnter keyCode), class "form-control", type_ "text", value str ]
        []


encodeTexts : Array String -> String
encodeTexts texts =
    let
        encodeParam param =
            percentEncode "text[]" ++ "=" ++ percentEncode param
    in
    texts |> Array.map encodeParam |> Array.toList |> String.join "&"



-- HTTP


getConfig : Cmd Msg
getConfig =
    Http.get
        { url = "/config.json"
        , expect = Http.expectJson GotConfig configDecoder
        }



-- JSON decode


configDecoder : D.Decoder (Dict String Meme)
configDecoder =
    D.dict memeDecoder


memeDecoder : D.Decoder Meme
memeDecoder =
    D.map6 Meme
        (D.field "id" D.string)
        (D.field "name" D.string)
        (D.field "filename" D.string)
        (D.field "font-color" D.string)
        (D.field "line-chars" D.int)
        (D.field "boxes" (D.list (D.list D.int)))
